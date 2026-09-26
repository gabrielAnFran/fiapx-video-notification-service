package config

import "testing"

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("NOTIFICATION_PORT", "")
	t.Setenv("NOTIFICATION_DB_DSN", "")
	t.Setenv("NOTIFICATION_AMQP_URL", "")
	t.Setenv("SMTP_HOST", "")
	t.Setenv("SMTP_PORT", "")
	t.Setenv("SMTP_USER", "")
	t.Setenv("SMTP_PASSWORD", "")
	t.Setenv("SMTP_FROM_EMAIL", "")

	cfg := Load()

	if cfg.Port != "8083" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8083")
	}
	if cfg.DBDSN != "host=localhost user=postgres password=postgres dbname=notification_service port=5432 sslmode=disable" {
		t.Errorf("unexpected default DBDSN: %q", cfg.DBDSN)
	}
	if cfg.AMQPURL != "amqp://guest:guest@localhost:5672/" { //nolint:gosec // default RabbitMQ dev credentials, not a real secret
		t.Errorf("unexpected default AMQPURL: %q", cfg.AMQPURL)
	}
	if cfg.SMTPHost != "" {
		t.Errorf("SMTPHost = %q, want empty", cfg.SMTPHost)
	}
	if cfg.SMTPPort != "" {
		t.Errorf("SMTPPort = %q, want empty", cfg.SMTPPort)
	}
	if cfg.SMTPUser != "" {
		t.Errorf("SMTPUser = %q, want empty", cfg.SMTPUser)
	}
	if cfg.SMTPPassword != "" {
		t.Errorf("SMTPPassword = %q, want empty", cfg.SMTPPassword)
	}
	if cfg.SMTPFrom != "no-reply@fiapx.local" {
		t.Errorf("SMTPFrom = %q, want %q", cfg.SMTPFrom, "no-reply@fiapx.local")
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("NOTIFICATION_PORT", "9999")
	t.Setenv("NOTIFICATION_DB_DSN", "custom-dsn")
	t.Setenv("NOTIFICATION_AMQP_URL", "amqp://custom/")
	t.Setenv("SMTP_HOST", "smtp.example.com")
	t.Setenv("SMTP_PORT", "587")
	t.Setenv("SMTP_USER", "user")
	t.Setenv("SMTP_PASSWORD", "pass")
	t.Setenv("SMTP_FROM_EMAIL", "notify@fiapx.com")

	cfg := Load()

	if cfg.Port != "9999" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9999")
	}
	if cfg.DBDSN != "custom-dsn" {
		t.Errorf("DBDSN = %q, want %q", cfg.DBDSN, "custom-dsn")
	}
	if cfg.AMQPURL != "amqp://custom/" {
		t.Errorf("AMQPURL = %q, want %q", cfg.AMQPURL, "amqp://custom/")
	}
	if cfg.SMTPHost != "smtp.example.com" {
		t.Errorf("SMTPHost = %q, want %q", cfg.SMTPHost, "smtp.example.com")
	}
	if cfg.SMTPPort != "587" {
		t.Errorf("SMTPPort = %q, want %q", cfg.SMTPPort, "587")
	}
	if cfg.SMTPUser != "user" {
		t.Errorf("SMTPUser = %q, want %q", cfg.SMTPUser, "user")
	}
	if cfg.SMTPPassword != "pass" {
		t.Errorf("SMTPPassword = %q, want %q", cfg.SMTPPassword, "pass")
	}
	if cfg.SMTPFrom != "notify@fiapx.com" {
		t.Errorf("SMTPFrom = %q, want %q", cfg.SMTPFrom, "notify@fiapx.com")
	}
}
