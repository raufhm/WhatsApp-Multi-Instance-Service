package main

import (
	"context"
	"log"
	"net/http"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	cfg "github.com/raufhm/whatsapp-testing/config"
	"github.com/raufhm/whatsapp-testing/handler"
	"github.com/raufhm/whatsapp-testing/internal/storage"
	"github.com/raufhm/whatsapp-testing/whatsapp"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	config, err := cfg.LoadConfig()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	pgStore, err := storage.NewPostgresStore(config.PGDSN)
	if err != nil {
		log.Fatalf("Store error: %v", err)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("AWS config error: %v", err)
	}
	s3Client := s3.NewFromConfig(awsCfg)
	s3Store := storage.NewS3Storage(s3Client, config.S3Bucket)

	dispatcher := whatsapp.NewMultiDispatcher(
		&whatsapp.LoggerDispatcher{},
		pgStore,
	)

	container, err := sqlstore.New(context.Background(), "pgx", config.PGDSN, waLog.Stdout("DB", "WARN", true))
	if err != nil {
		log.Fatalf("Container error: %v", err)
	}

	manager := whatsapp.NewWhatsAppManager(container, dispatcher, s3Store)
	if err := manager.Start(); err != nil {
		log.Printf("Manager start error: %v", err)
	}

	srv := &handler.Server{Manager: manager}

	http.HandleFunc("/api/onboard", srv.OnboardHandler)
	http.HandleFunc("/api/send", srv.SendHandler)
	http.HandleFunc("/api/bots", srv.ListBotsHandler)
	http.HandleFunc("/api/bots/detail", srv.GetBotHandler)
	http.HandleFunc("/api/health", handler.HealthHandler)

	log.Printf("Server starting on port %s...", config.Port)
	if err := http.ListenAndServe(":"+config.Port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
