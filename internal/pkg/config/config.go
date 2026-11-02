package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const minJWTSecretBytes = 32

var defaultJWTSecrets = []string{"changemechangemechangemechangeme12"}

type Config struct {
	HTTPAddr    string
	PostgresDSN string
	RedisAddr   string

	JWTSecret  string
	JWTTTL     time.Duration
	BcryptCost int

	S3Endpoint      string
	S3Region        string
	S3AccessKey     string
	S3SecretKey     string
	S3BucketPublic  string
	S3BucketPrivate string
	S3UsePathStyle  bool
}

func Load() (*Config, error) {
	cfg := &Config{
		HTTPAddr:    getEnv("APP_HTTP_ADDR", ":8080"),
		PostgresDSN: os.Getenv("APP_POSTGRES_DSN"),
		RedisAddr:   getEnv("APP_REDIS_ADDR", "localhost:6379"),

		JWTSecret: os.Getenv("APP_JWT_SECRET"),

		S3Endpoint:      getEnv("APP_S3_ENDPOINT", "http://localhost:8333"),
		S3Region:        getEnv("APP_S3_REGION", "us-east-1"),
		S3AccessKey:     os.Getenv("APP_S3_ACCESS_KEY"),
		S3SecretKey:     os.Getenv("APP_S3_SECRET_KEY"),
		S3BucketPublic:  getEnv("APP_S3_BUCKET_PUBLIC", "craftshelf-public"),
		S3BucketPrivate: getEnv("APP_S3_BUCKET_PRIVATE", "craftshelf-private"),
	}

	if cfg.PostgresDSN == "" {
		return nil, fmt.Errorf("APP_POSTGRES_DSN is required")
	}
	var err error

	if cfg.JWTSecret, err = validateJWTSecret(cfg.JWTSecret); err != nil {
		return nil, err
	}

	if cfg.JWTTTL, err = parseDuration(getEnv("APP_JWT_TTL", "24h"), "APP_JWT_TTL"); err != nil {
		return nil, err
	}
	if cfg.BcryptCost, err = parseIntInRange(getEnv("APP_BCRYPT_COST", "12"), "APP_BCRYPT_COST", 10, 14); err != nil {
		return nil, err
	}
	if cfg.S3UsePathStyle, err = parseBool(getEnv("APP_S3_USE_PATH_STYLE", "true"), "APP_S3_USE_PATH_STYLE"); err != nil {
		return nil, err
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func validateJWTSecret(raw string) (string, error) {
	secret := strings.TrimSpace(raw)
	if secret == "" {
		return "", errors.New("APP_JWT_SECRET is required")
	}
	if len(secret) < minJWTSecretBytes {
		return "", fmt.Errorf("APP_JWT_SECRET must be at least %d bytes (got %d)", minJWTSecretBytes, len(secret))
	}
	for _, r := range secret {
		if r <= ' ' || r == 0x7f {
			return "", errors.New("APP_JWT_SECRET must not contain whitespace or control characters")
		}
	}
	for _, placeholder := range defaultJWTSecrets {
		if secret == placeholder {
			return "", errors.New("APP_JWT_SECRET must not be a default placeholder")
		}
	}
	return secret, nil
}

func parseDuration(raw, key string) (time.Duration, error) {
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s %q: %w", key, raw, err)
	}
	return d, nil
}

func parseIntInRange(raw, key string, min, max int) (int, error) {
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s %q: %w", key, raw, err)
	}
	if n < min || n > max {
		return 0, fmt.Errorf("%s must be in [%d,%d], got %d", key, min, max, n)
	}
	return n, nil
}

func parseBool(raw, key string) (bool, error) {
	b, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("parse %s %q: %w", key, raw, err)
	}
	return b, nil
}
