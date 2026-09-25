package config

import (
	"os"
)

type Config struct {
	Port         string
	DBDSN        string
	AMQPURL      string
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
}

func Load() Config {
	return Config{
		Port:         getEnv("NOTIFICATION_PORT", "8083"),
		DBDSN:        getEnv("NOTIFICATION_DB_DSN", "host=localhost user=postgres password=postgres dbname=notification_service port=5432 sslmode=disable"),
		AMQPURL:      getEnv("NOTIFICATION_AMQP_URL", "amqp://guest:guest@localhost:5672/"),
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", ""),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM_EMAIL", "no-reply@fiapx.local"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
