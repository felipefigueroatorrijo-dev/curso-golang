package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func initMongoForTest(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := InitMongo(ctx); err != nil {
		t.Skipf("skipping DB connection test: %v", err)
	}
	if client == nil {
		t.Fatal("expected MongoDB client to be initialized")
	}
}

func TestDBConnection(t *testing.T) {
	initMongoForTest(t)

	ctxPing, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()
	if err := client.Ping(ctxPing, nil); err != nil {
		t.Fatalf("MongoDB ping failed: %v", err)
	}
}

func TestDBHealthHandler(t *testing.T) {
	initMongoForTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/db/health", nil)
	rr := httptest.NewRecorder()
	DBHealthHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if got := rr.Body.String(); got != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected body: %s", got)
	}
}

func TestLoadEnvFileSetsEnvVars(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "conn_test.env")
	content := "DB_HOST=localhost\nDB_PORT=27017\nDB_NAME=testdb\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_NAME")

	if err := loadEnvFile(path); err != nil {
		t.Fatalf("loadEnvFile failed: %v", err)
	}

	if got := os.Getenv("DB_NAME"); got != "testdb" {
		t.Fatalf("expected DB_NAME=testdb, got %q", got)
	}
}

func TestLoadEnvFileDoesNotOverrideExistingEnv(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "conn_test.env")
	content := "DB_NAME=fromfile\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	os.Setenv("DB_NAME", "existing")
	defer os.Unsetenv("DB_NAME")

	if err := loadEnvFile(path); err != nil {
		t.Fatalf("loadEnvFile failed: %v", err)
	}

	if got := os.Getenv("DB_NAME"); got != "existing" {
		t.Fatalf("expected DB_NAME=existing, got %q", got)
	}
}

func TestAtoiInvalid(t *testing.T) {
	if got := atoi("not-a-number"); got != 0 {
		t.Fatalf("expected atoi to return 0 for invalid input, got %d", got)
	}
}
