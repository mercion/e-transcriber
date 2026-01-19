package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	schema "github.com/mutablelogic/go-whisper/pkg/schema"
	whisper "github.com/mutablelogic/go-whisper/pkg/whisper"
)

func main() {
	cfg := loadConfig()

	if err := os.MkdirAll(cfg.ModelsDir, 0o755); err != nil {
		log.Fatalf("failed to create models dir: %v", err)
	}

	manager, err := whisper.New(cfg.ModelsDir)
	if err != nil {
		log.Fatalf("failed to init whisper manager: %v", err)
	}
	defer whisper.Close()

	model, err := ensureModel(context.Background(), manager, cfg.ModelID, cfg.ModelPath)
	if err != nil {
		log.Fatalf("failed to load model: %v", err)
	}

	server := NewServer(cfg, manager, model)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("shutting down")
		_ = whisper.Close()
		os.Exit(0)
	}()

	if err := server.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func ensureModel(ctx context.Context, manager *whisper.Manager, modelID, modelPath string) (*schema.Model, error) {
	if modelID != "" {
		if model := manager.GetModelById(modelID); model != nil {
			return model, nil
		}
	}

	if modelPath == "" {
		return nil, errMissingModel
	}

	model, err := manager.DownloadModel(ctx, modelPath, nil)
	if err != nil {
		return nil, err
	}
	return model, nil
}
