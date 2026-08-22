package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config holds all runtime settings resolved from environment variables.
// Every field has a documented default and a corresponding env key.
type Config struct {
	// PG_DSN — PostgreSQL data source name (required).
	PGDSN string
	// PORT — HTTP listen port (default: 8080).
	Port string
	// S3_BUCKET — S3-compatible bucket for outgoing media (optional; enables S3 storage).
	S3Bucket string
	// S3_ENDPOINT — custom S3-compatible endpoint (e.g. Cloudflare R2).
	S3Endpoint string
	// S3_REGION — region for the S3-compatible endpoint (R2 uses "auto").
	S3Region string
	// S3_OBJECT_URL — public/CDN base URL for S3 media objects (e.g. https://cdn.example.com or https://bucket.s3.region.amazonaws.com).
	S3ObjectURL string
	// MEDIA_DIR — local disk storage directory when S3 is not configured (default: ./media).
	MediaDir string
	// BOT_SESSION_TIMEOUT — idle session timeout before auto-closure (default: 30m).
	BotSessionTimeout time.Duration
	// BOT_FALLBACK_REPLY — message sent when no rule matches (default: empty, no reply).
	BotFallbackReply string
	// BOT_RULES_VERSION — version tag written to bot_sessions for audit (default: default).
	BotRulesVersion string
	// LOG_LEVEL — application log verbosity: DEBUG, INFO, WARN, ERROR (default: INFO).
	LogLevel string
	// UPLOAD_WORKER_ENABLED — enable the retryable S3 upload worker (default: true).
	UploadWorkerEnabled bool
	// UPLOAD_POLL_INTERVAL — worker poll interval (default: 5s).
	UploadPollInterval time.Duration
	// UPLOAD_MAX_ATTEMPTS — maximum upload attempts before permanent failure (default: 5).
	UploadMaxAttempts int
	// UPLOAD_INITIAL_BACKOFF — initial retry backoff (default: 1s).
	UploadInitialBackoff time.Duration
	// UPLOAD_MAX_BACKOFF — maximum retry backoff cap (default: 60s).
	UploadMaxBackoff time.Duration
	// UPLOAD_LEASE — claim lease duration; an expired lease lets another worker
	// reclaim a PROCESSING job after a crash (default: 60s).
	UploadLease time.Duration
	// UPLOAD_JITTER — random backoff jitter factor, 0 disables (default: 0.2).
	UploadJitter float64
}

func LoadConfig() (Config, error) {
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("MEDIA_DIR", "./media")
	viper.SetDefault("BOT_SESSION_TIMEOUT", "30m")
	viper.SetDefault("BOT_FALLBACK_REPLY", "")
	viper.SetDefault("BOT_RULES_VERSION", "default")
	viper.SetDefault("LOG_LEVEL", "INFO")
	viper.SetDefault("UPLOAD_WORKER_ENABLED", "true")
	viper.SetDefault("UPLOAD_POLL_INTERVAL", "5s")
	viper.SetDefault("UPLOAD_MAX_ATTEMPTS", "5")
	viper.SetDefault("UPLOAD_INITIAL_BACKOFF", "1s")
	viper.SetDefault("UPLOAD_MAX_BACKOFF", "60s")
	viper.SetDefault("UPLOAD_LEASE", "60s")
	viper.SetDefault("UPLOAD_JITTER", "0.2")
	viper.AutomaticEnv()

	timeout, err := time.ParseDuration(viper.GetString("BOT_SESSION_TIMEOUT"))
	if err != nil {
		return Config{}, err
	}
	pollInterval, err := time.ParseDuration(viper.GetString("UPLOAD_POLL_INTERVAL"))
	if err != nil {
		return Config{}, err
	}
	initialBackoff, err := time.ParseDuration(viper.GetString("UPLOAD_INITIAL_BACKOFF"))
	if err != nil {
		return Config{}, err
	}
	maxBackoff, err := time.ParseDuration(viper.GetString("UPLOAD_MAX_BACKOFF"))
	if err != nil {
		return Config{}, err
	}
	lease, err := time.ParseDuration(viper.GetString("UPLOAD_LEASE"))
	if err != nil {
		return Config{}, err
	}

	return Config{
		Port:                 viper.GetString("PORT"),
		PGDSN:                viper.GetString("PG_DSN"),
		S3Bucket:             viper.GetString("S3_BUCKET"),
		S3Endpoint:           viper.GetString("S3_ENDPOINT"),
		S3Region:             viper.GetString("S3_REGION"),
		S3ObjectURL:          viper.GetString("S3_OBJECT_URL"),
		MediaDir:             viper.GetString("MEDIA_DIR"),
		BotSessionTimeout:    timeout,
		BotFallbackReply:     viper.GetString("BOT_FALLBACK_REPLY"),
		BotRulesVersion:      viper.GetString("BOT_RULES_VERSION"),
		LogLevel:             viper.GetString("LOG_LEVEL"),
		UploadWorkerEnabled:  viper.GetBool("UPLOAD_WORKER_ENABLED"),
		UploadPollInterval:   pollInterval,
		UploadMaxAttempts:    viper.GetInt("UPLOAD_MAX_ATTEMPTS"),
		UploadInitialBackoff: initialBackoff,
		UploadMaxBackoff:     maxBackoff,
		UploadLease:          lease,
		UploadJitter:         viper.GetFloat64("UPLOAD_JITTER"),
	}, nil
}
