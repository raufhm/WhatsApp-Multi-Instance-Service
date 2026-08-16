package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	cfg "github.com/raufhm/whatsapp-testing/config"
	"github.com/raufhm/whatsapp-testing/handler"
	"github.com/raufhm/whatsapp-testing/internal/bot"
	"github.com/raufhm/whatsapp-testing/internal/storage"
	"github.com/raufhm/whatsapp-testing/internal/upload"
	"github.com/raufhm/whatsapp-testing/whatsapp"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	config, err := cfg.LoadConfig()
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	pgStore, err := storage.NewPostgresStore(config.PGDSN)
	if err != nil {
		log.Fatalf("Store error: %v", err)
	}
	pgStore.SetUploadClaimLease(config.UploadLease)

	var mediaStore storage.MediaStore
	var s3Store *storage.S3Storage

	if config.S3Bucket != "" {
		awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
		if err != nil {
			log.Fatalf("AWS config error: %v", err)
		}
		s3Client := s3.NewFromConfig(awsCfg)
		s3Store = storage.NewS3Storage(s3Client, config.S3Bucket)
		mediaStore = s3Store
	} else {
		diskStore, err := storage.NewDiskStore(config.MediaDir)
		if err != nil {
			log.Fatalf("Disk media store error: %v", err)
		}
		mediaStore = diskStore
	}

	// The retryable S3 upload worker is only started when archiving is enabled.
	// It is wired into the manager so outgoing media enqueues durable jobs whose
	// uploads are retried with bounded exponential backoff.
	var uploadManager *upload.Manager
	if config.S3Bucket != "" {
		uploadManager = upload.NewManager(pgStore, s3Store, upload.Config{
			Enabled:        config.UploadWorkerEnabled,
			PollInterval:   config.UploadPollInterval,
			MaxAttempts:    config.UploadMaxAttempts,
			InitialBackoff: config.UploadInitialBackoff,
			MaxBackoff:     config.UploadMaxBackoff,
			Jitter:         config.UploadJitter,
		}, func(hostID string) uuid.UUID {
			tenant, err := pgStore.TenantForHost(hostID)
			if err != nil {
				return uuid.Nil
			}
			return tenant
		})
	}

	botSender := whatsapp.NewBotSender()
	botProcessor := bot.NewProcessor(bot.NewEngine(bot.Config{
		Fallback:       config.BotFallbackReply,
		RuleVersion:    config.BotRulesVersion,
		SessionTimeout: config.BotSessionTimeout,
	}), botSender, pgStore)
	projector := whatsapp.NewAsyncProjectorWithBot(pgStore, botProcessor, 256)
	dispatcher := whatsapp.NewMultiDispatcher(
		&whatsapp.LoggerDispatcher{},
		pgStore,
		projector,
	)

	container, err := sqlstore.New(context.Background(), "pgx", config.PGDSN, waLog.Stdout("DB", "WARN", true))
	if err != nil {
		log.Fatalf("Container error: %v", err)
	}

	manager := whatsapp.NewWhatsAppManager(container, dispatcher, s3Store)
	manager.MediaStore = mediaStore
	manager.S3ObjectURL = config.S3ObjectURL
	manager.Store = pgStore
	manager.Pairing.SetStore(pgStore)
	manager.ResolveTenant = func(hostID string) uuid.UUID {
		tenant, err := pgStore.TenantForHost(hostID)
		if err != nil {
			return uuid.Nil
		}
		return tenant
	}
	if uploadManager != nil {
		manager.Uploader = uploadManager
		uploadManager.Start(ctx)
		defer uploadManager.Wait()
	}
	botSender.SetManager(manager)
	go runSessionTimeoutLoop(pgStore, config.BotSessionTimeout)
	if err := manager.Start(); err != nil {
		log.Printf("Manager start error: %v", err)
	}

	srv := &handler.Server{
		Manager:     manager,
		Platform:    pgStore,
		MediaStore:  mediaStore,
		S3ObjectURL: config.S3ObjectURL,
		Auth:        pgStore,
	}

	http.HandleFunc("/api/onboard", srv.OnboardHandler)
	http.HandleFunc("/api/send", srv.SendHandler)
	http.HandleFunc("/api/bots", srv.ListBotsHandler)
	http.HandleFunc("/api/bots/detail", srv.GetBotHandler)
	http.HandleFunc("/api/health", handler.HealthHandler)
	http.HandleFunc("/api/v1/", srv.APIHandler)
	http.HandleFunc("/api/v1", srv.APIHandler)

	// Operator dashboard (built frontend + session auth).
	dashboard := &handler.DashboardHandler{Auth: pgStore, WhatsApp: manager, StaticFS: dashboardStaticFS()}
	http.Handle("/dashboard/", dashboard)
	http.Handle("/dashboard", dashboard)

	// Protected dashboard API routes.
	dashboardAPI := &handler.DashboardAPIHandler{
		Platform:    pgStore,
		Manager:     manager,
		Store:       pgStore,
		Auth:        pgStore,
		MediaStore:  mediaStore,
		S3ObjectURL: config.S3ObjectURL,
	}
	http.Handle("/dashboard/api/accounts", handler.DashboardSessionMiddleware(pgStore, dashboardAPI))
	http.Handle("/dashboard/api/pairing", handler.DashboardSessionMiddleware(pgStore, dashboardAPI))
	http.Handle("/dashboard/api/pairing/", handler.DashboardSessionMiddleware(pgStore, dashboardAPI))
	http.Handle("/dashboard/api/bot-rules", handler.DashboardSessionMiddleware(pgStore, dashboardAPI))
	http.Handle("/dashboard/api/bot-rules/activate", handler.DashboardSessionMiddleware(pgStore, dashboardAPI))
	http.Handle("/dashboard/api/upload-jobs", handler.DashboardSessionMiddleware(pgStore, dashboardAPI))
	http.Handle("/dashboard/api/operators", handler.DashboardSessionMiddleware(pgStore, dashboardAPI))
	http.Handle("/dashboard/api/media", handler.DashboardSessionMiddleware(pgStore, dashboardAPI))
	http.Handle("/dashboard/api/media/", handler.DashboardSessionMiddleware(pgStore, dashboardAPI))

	// Redirect root to dashboard.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
	})

	log.Printf("[startup] port=%s bot_timeout=%s bot_rules=%s log_level=%s upload_worker=%t",
		config.Port, config.BotSessionTimeout, config.BotRulesVersion, config.LogLevel, uploadManager != nil && config.UploadWorkerEnabled)
	server := &http.Server{Addr: ":" + config.Port, Handler: nil}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}

func runSessionTimeoutLoop(store *storage.PostgresStore, timeout time.Duration) {
	if timeout <= 0 {
		log.Printf("[timeout-loop] disabled (BOT_SESSION_TIMEOUT=0)")
		return
	}
	log.Printf("[timeout-loop] started (interval=1m, session_timeout=%s)", timeout)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for now := range ticker.C {
		if err := store.CloseAllTimedOut(timeout, now); err != nil {
			log.Printf("[timeout-loop] ERROR closing timed-out sessions: %v", err)
		}
	}
}
