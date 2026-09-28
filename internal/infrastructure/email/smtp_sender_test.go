package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSend_NotConfigured_NoOp(t *testing.T) {
	s := NewSender("", "", "", "", "no-reply@fiapx.local")
	err := s.Send("user@example.com", Notification{
		VideoID:          "11111111-1111-1111-1111-111111111111",
		OriginalFilename: "movie.mp4",
		Type:             "COMPLETED",
	})
	require.NoError(t, err)
}

func TestSend_NotConfigured_MissingPortOnly_NoOp(t *testing.T) {
	s := NewSender("smtp.example.com", "", "", "", "no-reply@fiapx.local")
	err := s.Send("user@example.com", Notification{Type: "FAILED"})
	require.NoError(t, err)
}

func TestSend_ConnectionRefused_ReturnsError(t *testing.T) {
	// Port 1 on localhost has nothing listening, so the SMTP dial fails
	// fast and deterministically — exercises Send's/sendEmail's error path
	// without needing a real (or fake) SMTP server.
	s := NewSender("127.0.0.1", "1", "", "", "no-reply@fiapx.local")
	err := s.Send("user@example.com", Notification{Type: "COMPLETED", OriginalFilename: "x.mp4"})
	assert.Error(t, err)
}

func TestBuildSubject_Completed(t *testing.T) {
	n := Notification{OriginalFilename: "ferias-2026.mp4", Type: "COMPLETED"}
	subject := buildSubject(n)
	assert.Contains(t, subject, "Processamento de Vídeo Concluído")
	assert.Contains(t, subject, "ferias-2026.mp4")
}

func TestBuildSubject_Failed(t *testing.T) {
	n := Notification{OriginalFilename: "ferias-2026.mp4", Type: "FAILED"}
	subject := buildSubject(n)
	assert.Contains(t, subject, "Falha no Processamento do Vídeo")
	assert.Contains(t, subject, "ferias-2026.mp4")
}

func TestBuildBody_Completed(t *testing.T) {
	n := Notification{
		VideoID:          "11111111-1111-1111-1111-111111111111",
		OriginalFilename: "ferias-2026.mp4",
		Type:             "COMPLETED",
	}
	body := buildBody(n)
	assert.Contains(t, body, "ferias-2026.mp4")
	assert.Contains(t, body, "11111111-1111-1111-1111-111111111111")
	assert.Contains(t, body, "Concluído")
	assert.NotContains(t, body, "Ocorreu um erro")
}

func TestBuildBody_Failed(t *testing.T) {
	n := Notification{
		VideoID:          "11111111-1111-1111-1111-111111111111",
		OriginalFilename: "ferias-2026.mp4",
		Type:             "FAILED",
		ErrorMessage:     "codec not supported",
	}
	body := buildBody(n)
	assert.Contains(t, body, "ferias-2026.mp4")
	assert.Contains(t, body, "11111111-1111-1111-1111-111111111111")
	assert.Contains(t, body, "codec not supported")
	assert.Contains(t, body, "Falha")
}
