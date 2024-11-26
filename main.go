package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/go-redis/redis/v8"
)

// Initialize global Redis client and context
var ctx = context.Background()
var rdb *redis.Client

func main() {
	// Ensure Redis server is running
	ensureRedisServerRunning()

	// Initialize Redis client
	initRedis()

	// Set up HTTP handlers
	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		value := r.URL.Query().Get("value")
		if err := setValue(key, value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Key %s set successfully!", key)
	})

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		value, err := getValue(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Value for key %s: %s", key, value)
	})

	// Start the HTTP server
	log.Println("Starting server on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// Ensure Redis server is running, and start it if necessary
func ensureRedisServerRunning() {
	cmd := exec.Command("redis-cli", "ping")
	if err := cmd.Run(); err != nil {
		log.Println("Redis server is not running. Attempting to start Redis server...")

		redisServerCmd := exec.Command("redis-server")
		redisServerCmd.Stdout = os.Stdout
		redisServerCmd.Stderr = os.Stderr

		if err := redisServerCmd.Start(); err != nil {
			log.Fatalf("Failed to start Redis server: %v", err)
		}

		log.Println("Redis server started successfully. Waiting for it to be ready...")
		time.Sleep(2 * time.Second) // Allow some time for the server to initialize
	} else {
		log.Println("Redis server is already running.")
	}
}

// Initialize Redis client
func initRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "",               // No password set
		DB:       0,                // Use default DB
	})

	// Test connection
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	fmt.Println("Connected to Redis!")
}

// Set a value in Redis
func setValue(key string, value string) error {
	err := rdb.Set(ctx, key, value, 0).Err()
	if err != nil {
		return fmt.Errorf("could not set key: %w", err)
	}
	return nil
}

// Get a value from Redis
func getValue(key string) (string, error) {
	val, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key does not exist")
	} else if err != nil {
		return "", fmt.Errorf("could not get key: %w", err)
	}
	return val, nil
}
