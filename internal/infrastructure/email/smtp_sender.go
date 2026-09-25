package email

import (
	"fmt"
	"log/slog"
	"net/smtp"
)

// Sender sends notification emails over SMTP. When host/port are not
// configured it degrades to a logging-only no-op, which is convenient for
// local development without real SMTP credentials. Point it at a local
// Mailhog instance (typically host=localhost, port=1025, no user/password)
// for a real end-to-end send during local dev: Mailhog accepts
// unauthenticated SMTP, so Sender skips smtp.PlainAuth entirely whenever
// both user and password are empty.
type Sender struct {
	host, port, user, password, from string
}

func NewSender(host, port, user, password, from string) *Sender {
	return &Sender{host: host, port: port, user: user, password: password, from: from}
}

// Notification carries the data needed to build the subject/body of a
// video-processing outcome email.
type Notification struct {
	VideoID          string
	OriginalFilename string
	Type             string // "COMPLETED" or "FAILED"
	ErrorMessage     string // only set when Type == "FAILED"
}

// buildSubject and buildBody are pure functions extracted purely so they're
// unit-testable without touching the network.

func buildSubject(n Notification) string {
	if n.Type == "FAILED" {
		return fmt.Sprintf("Falha no Processamento do Vídeo: %s", n.OriginalFilename)
	}
	return fmt.Sprintf("Processamento de Vídeo Concluído: %s", n.OriginalFilename)
}

func buildBody(n Notification) string {
	if n.Type == "FAILED" {
		return fmt.Sprintf(`Falha no Processamento do Vídeo

Arquivo: %s
ID do Vídeo: %s

Ocorreu um erro durante o processamento do seu vídeo:
%s

Esta é uma mensagem automática, por favor não responda.
`, n.OriginalFilename, n.VideoID, n.ErrorMessage)
	}
	return fmt.Sprintf(`Processamento de Vídeo Concluído

Arquivo: %s
ID do Vídeo: %s

Seu vídeo foi processado com sucesso. O arquivo zip com os frames extraídos
já está disponível para download no sistema.

Esta é uma mensagem automática, por favor não responda.
`, n.OriginalFilename, n.VideoID)
}

// Send builds a subject/body from the notification and emails it to `to`.
// It returns nil without erroring (logging "SMTP not configured, skipping")
// when host or port is empty, so the worker can run in environments without
// real SMTP credentials.
func (s *Sender) Send(to string, n Notification) error {
	if s.host == "" || s.port == "" {
		slog.Info("SMTP not configured, skipping", "to", to, "type", n.Type, "video_id", n.VideoID)
		return nil
	}

	if err := s.sendEmail(to, buildSubject(n), buildBody(n)); err != nil {
		return err
	}
	slog.Info("email sent", "to", to, "type", n.Type, "video_id", n.VideoID)
	return nil
}

func (s *Sender) sendEmail(to, subject, body string) error {
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", s.from, to, subject, body))

	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	var auth smtp.Auth
	if s.user != "" || s.password != "" {
		auth = smtp.PlainAuth("", s.user, s.password, s.host)
	}
	// else: leave auth nil — Mailhog and similar local dev SMTP relays accept
	// unauthenticated connections.

	if err := smtp.SendMail(addr, auth, s.from, []string{to}, msg); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}
