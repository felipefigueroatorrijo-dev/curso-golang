package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Conectar a MongoDB
	if err := InitMongo(ctx); err != nil {
		log.Fatalf("failed to init mongo: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/search", SearchHandler)
	mux.HandleFunc("/api/games/", GameDetailHandler)
	mux.HandleFunc("/api/db/health", DBHealthHandler)

	mux.HandleFunc("/api/library", LibraryHandler)
	mux.HandleFunc("/api/library/", LibraryItemHandler)
	mux.HandleFunc("/api/library/stats", LibraryStatsHandler)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("Server listening on :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
