package main

import (
	"backend-diplomski/services/upload-service/internal/config"
	"backend-diplomski/services/upload-service/internal/storage"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	minioStorage, err := storage.NewMinioStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize bucket")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := minioStorage.EnsureBucket(ctx); err != nil {
		log.Fatalf("Failed to ensure bucket: %V", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("upload-service ok"))
	})

	mux.HandleFunc("/internal/test-upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()
		objectKey, err := minioStorage.UploadTestFile(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		response := map[string]string{
			"message":   "test file uploaded successfully",
			"bucket":    cfg.S3Bucket,
			"objectKey": objectKey,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	})

	log.Printf("upload-service listening on port %s", cfg.AppPort)

	if err := http.ListenAndServe(":"+cfg.AppPort, mux); err != nil {
		log.Fatal(err)
	}
}
