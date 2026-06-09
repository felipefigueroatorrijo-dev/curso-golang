package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var RAWG_KEY = "945d345a57cc4c3fb7b4f67211edd4c8"
var RAWG_BASE = "https://api.rawg.io/api"

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError envía un error en formato JSON consistente y loggea el error interno si existe.
func writeError(w http.ResponseWriter, httpCode int, codeStr, msg string, internalErr error) {
	w.Header().Set("Content-Type", "application/json")
	if internalErr != nil {
		log.Println(internalErr)
	}
	if httpCode == http.StatusNoContent {
		w.WriteHeader(httpCode)
		return
	}
	w.WriteHeader(httpCode)
	_ = json.NewEncoder(w).Encode(map[string]string{"code": codeStr, "error": msg})
}

// --- RAWG proxy handlers ---
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "missing q param", nil)
		return
	}
	url := fmt.Sprintf("%s/games?key=%s&search=%s", RAWG_BASE, RAWG_KEY, q)
	resp, err := http.Get(url)
	if err != nil {
		writeError(w, http.StatusBadGateway, "bad_gateway", "RAWG service unavailable", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	// Decode minimal fields
	var raw struct {
		Results []struct {
			ID     int    `json:"id"`
			Name   string `json:"name"`
			Genres []struct {
				Name string `json:"name"`
			} `json:"genres"`
			Platforms []struct {
				Platform struct {
					Name string `json:"name"`
				} `json:"platform"`
			} `json:"platforms"`
			BackgroundImage string  `json:"background_image"`
			Rating          float64 `json:"rating"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		writeError(w, http.StatusBadGateway, "bad_gateway", "failed to parse RAWG response", err)
		return
	}
	out := make([]map[string]interface{}, 0, len(raw.Results))
	for _, it := range raw.Results {
		genres := make([]string, 0, len(it.Genres))
		for _, g := range it.Genres {
			genres = append(genres, g.Name)
		}
		plats := make([]string, 0, len(it.Platforms))
		for _, p := range it.Platforms {
			plats = append(plats, p.Platform.Name)
		}
		out = append(out, map[string]interface{}{
			"id":        it.ID,
			"name":      it.Name,
			"genres":    genres,
			"platforms": plats,
			"image":     it.BackgroundImage,
			"rating":    it.Rating,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

func GameDetailHandler(w http.ResponseWriter, r *http.Request) {
	// path: /api/games/{id}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		writeError(w, http.StatusBadRequest, "bad_request", "missing id", nil)
		return
	}
	id := parts[2]
	url := fmt.Sprintf("%s/games/%s?key=%s", RAWG_BASE, id, RAWG_KEY)
	resp, err := http.Get(url)
	if err != nil {
		writeError(w, http.StatusBadGateway, "bad_gateway", "RAWG service unavailable", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		writeError(w, http.StatusBadGateway, "bad_gateway", "failed to parse RAWG respuesta", err)
		return
	}
	writeJSON(w, http.StatusOK, raw)
}

func DBHealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
		return
	}

	if client == nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "database client not initialized", nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "database connection failed", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- Library handlers (MongoDB) ---

type collectionInterface interface {
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error)
	CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error)
	InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
	UpdateByID(ctx context.Context, id interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
	FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult
	DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error)
}

var getCollection = func() collectionInterface {
	return client.Database(dbName).Collection(collectionName)
}

func LibraryHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		status := r.URL.Query().Get("status")
		filter := bson.M{}
		if status != "" {
			filter["status"] = status
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cur, err := getCollection().Find(ctx, filter)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "database error", err)
			return
		}
		defer cur.Close(ctx)
		var out []Game
		for cur.Next(ctx) {
			var g Game
			_ = cur.Decode(&g)
			out = append(out, g)
		}
		writeJSON(w, http.StatusOK, out)
		return
	case http.MethodPost:
		var payload struct {
			RawgID   int    `json:"rawg_id"`
			Title    string `json:"title"`
			Genre    string `json:"genre"`
			Platform string `json:"platform"`
			CoverURL string `json:"cover_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid json", err)
			return
		}
		if payload.RawgID == 0 || payload.Title == "" {
			writeError(w, http.StatusBadRequest, "bad_request", "rawg_id and title required", nil)
			return
		}
		// check duplicate
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		col := getCollection()
		cnt, err := col.CountDocuments(ctx, bson.M{"rawg_id": payload.RawgID})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "database error", err)
			return
		}
		if cnt > 0 {
			writeError(w, http.StatusConflict, "conflict", "rawg_id already exists", nil)
			return
		}
		now := time.Now()
		g := Game{
			RawgID:   payload.RawgID,
			Title:    payload.Title,
			Genre:    payload.Genre,
			Platform: payload.Platform,
			CoverURL: payload.CoverURL,
			AddedAt:  now,
		}
		res, err := col.InsertOne(ctx, g)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "failed to insert document", err)
			return
		}
		g.ID = res.InsertedID.(primitive.ObjectID)
		writeJSON(w, http.StatusCreated, g)
		return
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
		return
	}
}

func LibraryItemHandler(w http.ResponseWriter, r *http.Request) {
	// routes: /api/library/{id} for PUT and DELETE
	path := strings.TrimPrefix(r.URL.Path, "/api/library/")
	if path == "" || path == "/" || strings.Contains(path, "/") {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id", nil)
		return
	}
	idStr := strings.Trim(path, "/")
	oid, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid id format", err)
		return
	}
	col := getCollection()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	switch r.Method {
	case http.MethodPut:
		var payload struct {
			PersonalNote  *string `json:"personal_note"`
			PersonalScore *int    `json:"personal_score"`
			Status        *string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "bad_request", "invalid json", err)
			return
		}
		update := bson.M{}
		set := bson.M{}
		if payload.PersonalNote != nil {
			set["personal_note"] = *payload.PersonalNote
		}
		if payload.PersonalScore != nil {
			if *payload.PersonalScore < 1 || *payload.PersonalScore > 10 {
				writeError(w, http.StatusBadRequest, "bad_request", "personal_score must be 1-10", nil)
				return
			}
			set["personal_score"] = *payload.PersonalScore
		}
		if payload.Status != nil {
			if !allowedStatus[*payload.Status] {
				writeError(w, http.StatusBadRequest, "bad_request", "invalid status", nil)
				return
			}
			set["status"] = *payload.Status
		}
		if len(set) == 0 {
			writeError(w, http.StatusBadRequest, "bad_request", "no fields to update", nil)
			return
		}
		update["$set"] = set
		if _, err := col.UpdateByID(ctx, oid, update); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "database error", err)
			return
		}
		var g Game
		if err := col.FindOne(ctx, bson.M{"_id": oid}).Decode(&g); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "database error", err)
			return
		}
		writeJSON(w, http.StatusOK, g)
		return
	case http.MethodDelete:
		if _, err := col.DeleteOne(ctx, bson.M{"_id": oid}); err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "database error", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNoContent)
		return
	default:
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
		return
	}
}

func LibraryStatsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed", nil)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	col := getCollection()
	cur, err := col.Find(ctx, bson.M{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "database error", err)
		return
	}
	defer cur.Close(ctx)
	total := 0
	counts := map[string]int{"completado": 0, "jugando": 0, "pendiente": 0, "abandonado": 0}
	var sum int
	var scored int
	for cur.Next(ctx) {
		var g Game
		_ = cur.Decode(&g)
		total++
		if g.Status != "" {
			if _, ok := counts[g.Status]; ok {
				counts[g.Status]++
			}
		}
		if g.PersonalScore != nil {
			sum += *g.PersonalScore
			scored++
		}
	}
	avg := 0.0
	if scored > 0 {
		avg = float64(sum) / float64(scored)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total":         total,
		"by_status":     counts,
		"average_score": avg,
	})
}

// helper to parse int from string
func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
