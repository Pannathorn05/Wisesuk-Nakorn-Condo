package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env  string
	Port string

	DatabaseURL     string
	DBMaxConns      int32
	DBMinConns      int32
	DBConnectTimout time.Duration

	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration

	UploadDir       string
	PublicBaseURL   string
	MaxUploadBytes  int64
	AllowedOrigins  []string
	BcryptCost      int
	SeedAdminSecret string
}

// Load อ่านค่าจาก .env (ถ้ามี) แล้วทับด้วย environment variable จริง
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Env:             getEnv("APP_ENV", "development"),
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		DBMaxConns:      int32(getEnvInt("DB_MAX_CONNS", 10)),
		DBMinConns:      int32(getEnvInt("DB_MIN_CONNS", 2)),
		DBConnectTimout: time.Duration(getEnvInt("DB_CONNECT_TIMEOUT_SEC", 10)) * time.Second,
		JWTSecret:       getEnv("JWT_SECRET", ""),
		AccessTokenTTL:  time.Duration(getEnvInt("ACCESS_TOKEN_TTL_MIN", 30)) * time.Minute,
		RefreshTokenTTL: time.Duration(getEnvInt("REFRESH_TOKEN_TTL_DAY", 30)) * 24 * time.Hour,
		UploadDir:       getEnv("UPLOAD_DIR", "./uploads"),
		PublicBaseURL:   strings.TrimRight(getEnv("PUBLIC_BASE_URL", "http://localhost:8080"), "/"),
		MaxUploadBytes:  int64(getEnvInt("MAX_UPLOAD_MB", 5)) * 1024 * 1024,
		AllowedOrigins:  splitAndTrim(getEnv("ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")),
		BcryptCost:      getEnvInt("BCRYPT_COST", 12),
		SeedAdminSecret: getEnv("SEED_DEFAULT_PASSWORD", "Wisetsuk!2026"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("config: ต้องกำหนด DATABASE_URL")
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("config: ต้องกำหนด JWT_SECRET ความยาวอย่างน้อย 32 ตัวอักษร")
	}
	return cfg, nil
}

func (c *Config) IsProduction() bool { return c.Env == "production" }

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
