package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr              string
	LogLevel              string
	AllowedOrigins        []string
	DatabaseURL           string
	RabbitMQURL           string
	BaselineBaseURL       string
	BaselineTimeout       time.Duration
	BaselineAlgorithm     string
	BaselineReference     string
	KESMIBaseURL          string
	KESMITimeout          time.Duration
	KESMIModelID          string
	KESMIMaxRetries       int32
	KESMIRetryBackoff     time.Duration
	MigrationsDir         string
	JWTIssuer             string
	JWTAccessSecret       string
	JWTAccessTTL          time.Duration
	InitialUserLogin      string
	InitialUserPassword   string
	InitialUserRole       string
	S3Endpoint            string
	S3AccessKeyID         string
	S3SecretAccessKey     string
	S3Bucket              string
	S3UseSSL              bool
	MaxUploadSizeBytes    int64
	OutboxPollInterval    time.Duration
	OutboxMaxAttempts     int32
	AudioRetentionTTLDays int32
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:              envOrDefault("HTTP_ADDR", ":8080"),
		LogLevel:              envOrDefault("LOG_LEVEL", "INFO"),
		AllowedOrigins:        envCSVOrDefault("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://127.0.0.1:3000"}),
		MigrationsDir:         envOrDefault("MIGRATIONS_DIR", "migrations"),
		JWTIssuer:             envOrDefault("JWT_ISSUER", "core-backend"),
		InitialUserRole:       envOrDefault("INITIAL_USER_ROLE", "operator"),
		S3UseSSL:              envBoolOrDefault("MINIO_USE_SSL", false),
		MaxUploadSizeBytes:    envInt64OrDefault("MAX_UPLOAD_SIZE_BYTES", 25<<20),
		RabbitMQURL:           envOrDefault("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		BaselineBaseURL:       envOrDefault("BASELINE_SERVICE_URL", "http://ml-baseline:8080"),
		BaselineAlgorithm:     envOrDefault("BASELINE_ALGORITHM_VERSION", "baseline-v1"),
		BaselineReference:     envOrDefault("BASELINE_GENERAL_REFERENCE_VERSION", "general-v1"),
		KESMIBaseURL:          envOrDefault("KESMI_BASE_URL", "http://wimi:8081"),
		KESMIModelID:          envOrDefault("KESMI_MODEL_ID", "placeholder-model"),
		KESMIMaxRetries:       int32(envInt64OrDefault("KESMI_MAX_RETRIES", 2)),
		OutboxMaxAttempts:     int32(envInt64OrDefault("PROCESSING_OUTBOX_MAX_ATTEMPTS", 3)),
		AudioRetentionTTLDays: int32(envInt64OrDefault("AUDIO_RETENTION_TTL_DAYS", 30)),
	}

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	cfg.JWTAccessSecret = os.Getenv("JWT_ACCESS_SECRET")
	if cfg.JWTAccessSecret == "" {
		return Config{}, fmt.Errorf("JWT_ACCESS_SECRET is required")
	}
	var err error
	cfg.JWTAccessTTL, err = envDurationOrDefault("JWT_ACCESS_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, fmt.Errorf("JWT_ACCESS_TTL: %w", err)
	}
	cfg.OutboxPollInterval, err = envDurationOrDefault("PROCESSING_OUTBOX_POLL_INTERVAL", time.Second)
	if err != nil {
		return Config{}, fmt.Errorf("PROCESSING_OUTBOX_POLL_INTERVAL: %w", err)
	}
	cfg.BaselineTimeout, err = envDurationOrDefault("BASELINE_SERVICE_TIMEOUT", 3*time.Second)
	if err != nil {
		return Config{}, fmt.Errorf("BASELINE_SERVICE_TIMEOUT: %w", err)
	}
	cfg.KESMITimeout, err = envDurationOrDefault("KESMI_TIMEOUT_MS", 3*time.Second)
	if err != nil {
		return Config{}, fmt.Errorf("KESMI_TIMEOUT_MS: %w", err)
	}
	cfg.KESMIRetryBackoff, err = envDurationOrDefault("KESMI_RETRY_BACKOFF_MS", time.Second)
	if err != nil {
		return Config{}, fmt.Errorf("KESMI_RETRY_BACKOFF_MS: %w", err)
	}
	cfg.InitialUserLogin = os.Getenv("INITIAL_USER_LOGIN")
	cfg.InitialUserPassword = os.Getenv("INITIAL_USER_PASSWORD")
	cfg.S3Endpoint = os.Getenv("MINIO_ENDPOINT")
	if cfg.S3Endpoint == "" {
		return Config{}, fmt.Errorf("MINIO_ENDPOINT is required")
	}
	cfg.S3AccessKeyID = envFirst("MINIO_ACCESS_KEY_ID", "MINIO_ROOT_USER")
	if cfg.S3AccessKeyID == "" {
		return Config{}, fmt.Errorf("MINIO_ACCESS_KEY_ID or MINIO_ROOT_USER is required")
	}
	cfg.S3SecretAccessKey = envFirst("MINIO_SECRET_ACCESS_KEY", "MINIO_ROOT_PASSWORD")
	if cfg.S3SecretAccessKey == "" {
		return Config{}, fmt.Errorf("MINIO_SECRET_ACCESS_KEY or MINIO_ROOT_PASSWORD is required")
	}
	cfg.S3Bucket = os.Getenv("MINIO_BUCKET")
	if cfg.S3Bucket == "" {
		return Config{}, fmt.Errorf("MINIO_BUCKET is required")
	}
	if cfg.OutboxMaxAttempts <= 0 {
		return Config{}, fmt.Errorf("PROCESSING_OUTBOX_MAX_ATTEMPTS must be positive")
	}
	if cfg.KESMIMaxRetries <= 0 {
		return Config{}, fmt.Errorf("KESMI_MAX_RETRIES must be positive")
	}
	if cfg.AudioRetentionTTLDays <= 0 {
		return Config{}, fmt.Errorf("AUDIO_RETENTION_TTL_DAYS must be positive")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func envFirst(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}

func envDurationOrDefault(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	return time.ParseDuration(value)
}

func envBoolOrDefault(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt64OrDefault(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func envCSVOrDefault(key string, fallback []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}

	if len(result) == 0 {
		return fallback
	}

	return result
}
