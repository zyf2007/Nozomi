package server

import (
	"os"
	"path/filepath"
	"strings"
)

func loadSettings() Settings {
	dataDir := env("NOZOMI_DATA_DIR", "data")
	if !filepath.IsAbs(dataDir) {
		dataDir, _ = filepath.Abs(dataDir)
	}
	dbPath := env("NOZOMI_DB_PATH", filepath.Join(dataDir, "nozomi.sqlite3"))
	return Settings{
		HTTPAddr:      env("NOZOMI_HTTP_ADDR", "0.0.0.0:5000"),
		SMTPAddr:      env("NOZOMI_SMTP_ADDR", "0.0.0.0:2525"),
		DataDir:       dataDir,
		DBPath:        dbPath,
		CORSOrigins:   envList("NOZOMI_CORS_ORIGINS", []string{"http://localhost:5173", "http://127.0.0.1:5173"}),
		AdminUsername: env("NOZOMI_ADMIN_USERNAME", "admin"),
		AdminPassword: env("NOZOMI_ADMIN_PASSWORD", "change-me"),
		SessionSecret: env("NOZOMI_SESSION_SECRET", "change-this-session-secret"),
	}
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envList(key string, fallback []string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
