package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearchHandlerMissingQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/search", nil)
	rr := httptest.NewRecorder()

	SearchHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestGameDetailHandlerMissingID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/games/", nil)
	rr := httptest.NewRecorder()

	GameDetailHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestDBHealthHandlerMethodNotAllowed(t *testing.T) {
	client = nil
	req := httptest.NewRequest(http.MethodPost, "/api/db/health", nil)
	rr := httptest.NewRecorder()

	DBHealthHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestLibraryHandlerPostInvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/library", strings.NewReader("{invalid json"))
	rr := httptest.NewRecorder()

	LibraryHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestLibraryHandlerPostMissingFields(t *testing.T) {
	payload := `{"title":"Juego sin rawg_id"}`
	req := httptest.NewRequest(http.MethodPost, "/api/library", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	LibraryHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestLibraryItemHandlerInvalidPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/library/", strings.NewReader(`{"status":"jugando"}`))
	rr := httptest.NewRecorder()

	LibraryItemHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestLibraryItemHandlerInvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/library/12345", strings.NewReader(`{"status":"jugando"}`))
	rr := httptest.NewRecorder()

	LibraryItemHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestLibraryStatsHandlerMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/library/stats", nil)
	rr := httptest.NewRecorder()

	LibraryStatsHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestErrorResponseJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/search", nil)
	rr := httptest.NewRecorder()

	SearchHandler(rr, req)

	if rr.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", rr.Header().Get("Content-Type"))
	}

	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["code"] != "bad_request" {
		t.Fatalf("expected code bad_request, got %s", body["code"])
	}
}

func TestSearchHandlerSuccess(t *testing.T) {
	oldBase, oldKey := RAWG_BASE, RAWG_KEY
	defer func() {
		RAWG_BASE = oldBase
		RAWG_KEY = oldKey
	}()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/games") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"id":123,"name":"Zelda","genres":[{"name":"Adventure"}],"platforms":[{"platform":{"name":"Switch"}}],"background_image":"","rating":4.5}]}`))
	}))
	defer srv.Close()

	RAWG_BASE = srv.URL
	RAWG_KEY = "test"

	req := httptest.NewRequest(http.MethodGet, "/api/search?q=zelda", nil)
	rr := httptest.NewRecorder()

	SearchHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var out []map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(out) != 1 || out[0]["name"] != "Zelda" {
		t.Fatalf("unexpected response: %#v", out)
	}
}

func TestGameDetailHandlerSuccess(t *testing.T) {
	oldBase, oldKey := RAWG_BASE, RAWG_KEY
	defer func() {
		RAWG_BASE = oldBase
		RAWG_KEY = oldKey
	}()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/games/123") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":123,"name":"Zelda","description":"A game"}`))
	}))
	defer srv.Close()

	RAWG_BASE = srv.URL
	RAWG_KEY = "test"

	req := httptest.NewRequest(http.MethodGet, "/api/games/123", nil)
	rr := httptest.NewRecorder()

	GameDetailHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["name"] != "Zelda" {
		t.Fatalf("unexpected response: %#v", body)
	}
}

func TestWriteJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	writeJSON(rr, http.StatusCreated, map[string]string{"ok": "true"})

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}
	if rr.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %s", rr.Header().Get("Content-Type"))
	}
	if body, _ := io.ReadAll(rr.Body); !strings.Contains(string(body), "ok") {
		t.Fatalf("expected body to contain ok, got %s", body)
	}
}

func TestWriteErrorNoContent(t *testing.T) {
	rr := httptest.NewRecorder()
	writeError(rr, http.StatusNoContent, "no_content", "", nil)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rr.Code)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	called := false
	handler := loggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !called {
		t.Fatal("expected next handler to be called")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestLibraryHandlerMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/library", nil)
	rr := httptest.NewRecorder()

	LibraryHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}
