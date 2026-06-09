package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var dbName = "goVideoJuegos"
var collectionName = "game_library"

// InitMongo inicializa la conexión a MongoDB.
// Primero carga valores desde conn.env si existe. Luego toma variables de entorno.
func InitMongo(ctx context.Context) error {
	if err := loadEnvFile("conn.env"); err != nil {
		return err
	}

	if env := os.Getenv("DB_NAME"); env != "" {
		dbName = env
	}

	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = os.Getenv("MONGO_HOST")
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "27017"
	}

	if host == "" {
		host = "localhost"
	}

	hostPort := host
	if !strings.Contains(host, ":") {
		hostPort = fmt.Sprintf("%s:%s", host, port)
	}

	// Build URI without embedding credentials; provide credentials via options.Credential
	uri := fmt.Sprintf("mongodb://%s", hostPort)

	opts := options.Client().ApplyURI(uri)
	if user != "" {
		creds := options.Credential{
			Username: user,
			Password: pass,
		}
		// Allow overriding auth source/mechanism via env vars, otherwise use DB name
		if authSource := os.Getenv("DB_AUTH_SOURCE"); authSource != "" {
			creds.AuthSource = authSource
		} else {
			creds.AuthSource = dbName
		}
		if mech := os.Getenv("DB_AUTH_MECHANISM"); mech != "" {
			creds.AuthMechanism = mech
		}
		opts.SetAuth(creds)
	}

	c, err := mongo.Connect(ctx, opts)
	if err != nil {
		return err
	}

	// Ping
	ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.Ping(ctx2, nil); err != nil {
		return err
	}
	client = c

	// Verificar y crear la colección si no existe
	if err := ensureGameCollectionExists(ctx); err != nil {
		return err
	}

	return nil
}

func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

// ensureGameCollectionExists verifica si la colección game_library existe.
// Si no existe, la crea.
func ensureGameCollectionExists(ctx context.Context) error {
	db := client.Database(dbName)

	// Listar colecciones existentes
	collections, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("error listing collections: %w", err)
	}

	// Verificar si la colección ya existe
	for _, coll := range collections {
		if coll == collectionName {
			fmt.Printf("Colección '%s' ya existe\n", collectionName)
			return nil
		}
	}

	// Crear la colección si no existe
	err = db.CreateCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("error creating collection '%s': %w", collectionName, err)
	}

	fmt.Printf("Colección '%s' creada exitosamente\n", collectionName)
	return nil
}
