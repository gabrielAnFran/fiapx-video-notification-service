//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gabrielAnFran/fiapx-video-notification-service/internal/infrastructure/email"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mailhogMessage mirrors the subset of Mailhog's v2 API message shape
// (GET /api/v2/messages) that these tests care about.
type mailhogMessage struct {
	To []struct {
		Mailbox string `json:"Mailbox"`
		Domain  string `json:"Domain"`
	} `json:"To"`
	Content struct {
		Headers map[string][]string `json:"Headers"`
		Body    string              `json:"Body"`
	} `json:"Content"`
}

type mailhogMessagesResponse struct {
	Total int              `json:"total"`
	Items []mailhogMessage `json:"items"`
}

// fetchMailhogMessages polls Mailhog's HTTP API until it has at least one
// message (or the deadline elapses) and returns whatever it last saw.
func fetchMailhogMessages(t *testing.T) mailhogMessagesResponse {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	var last mailhogMessagesResponse
	for time.Now().Before(deadline) {
		resp, err := http.Get(testMailhogAPI + "/api/v2/messages")
		if err == nil {
			func() {
				defer resp.Body.Close()
				var parsed mailhogMessagesResponse
				if decErr := json.NewDecoder(resp.Body).Decode(&parsed); decErr == nil {
					last = parsed
				}
			}()
			if last.Total > 0 {
				return last
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return last
}

// clearMailhog deletes all messages from Mailhog so each test starts from a
// clean inbox despite sharing one Mailhog container across the package.
func clearMailhog(t *testing.T) {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, testMailhogAPI+"/api/v1/messages", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
}

func TestEmailSender_Send_DeliversToMailhog_Completed(t *testing.T) {
	clearMailhog(t)

	sender := email.NewSender(testSMTPHost, testSMTPPort, "", "", "no-reply@fiapx.local")
	to := "user@example.com"

	err := sender.Send(to, email.Notification{
		VideoID:          "11111111-1111-1111-1111-111111111111",
		OriginalFilename: "meu-video.mp4",
		Type:             "COMPLETED",
	})
	require.NoError(t, err)

	msgs := fetchMailhogMessages(t)
	require.Equal(t, 1, msgs.Total, "expected exactly one message delivered to mailhog")

	got := msgs.Items[0]
	require.Len(t, got.To, 1)
	assert.Equal(t, "user", got.To[0].Mailbox)
	assert.Equal(t, "example.com", got.To[0].Domain)

	subjects := got.Content.Headers["Subject"]
	require.NotEmpty(t, subjects)
	assert.Contains(t, subjects[0], "meu-video.mp4")
	assert.Contains(t, subjects[0], "Concluído")
}

func TestEmailSender_Send_DeliversToMailhog_Failed(t *testing.T) {
	clearMailhog(t)

	sender := email.NewSender(testSMTPHost, testSMTPPort, "", "", "no-reply@fiapx.local")
	to := "other@example.com"

	err := sender.Send(to, email.Notification{
		VideoID:          "22222222-2222-2222-2222-222222222222",
		OriginalFilename: "outro-video.mp4",
		Type:             "FAILED",
		ErrorMessage:     "codec not supported",
	})
	require.NoError(t, err)

	msgs := fetchMailhogMessages(t)
	require.Equal(t, 1, msgs.Total, "expected exactly one message delivered to mailhog")

	got := msgs.Items[0]
	require.Len(t, got.To, 1)
	assert.Equal(t, "other", got.To[0].Mailbox)
	assert.Equal(t, "example.com", got.To[0].Domain)

	subjects := got.Content.Headers["Subject"]
	require.NotEmpty(t, subjects)
	assert.Contains(t, subjects[0], "outro-video.mp4")
	assert.Contains(t, subjects[0], "Falha")
	assert.Contains(t, got.Content.Body, "codec not supported")
}

func TestEmailSender_Send_NotConfigured_SkipsWithoutError(t *testing.T) {
	// Empty host/port makes Sender degrade to a logging-only no-op — this
	// must never attempt to reach Mailhog (or fail) even though a real
	// SMTP server is available in this test suite.
	sender := email.NewSender("", "", "", "", "no-reply@fiapx.local")
	err := sender.Send("someone@example.com", email.Notification{
		VideoID:          "33333333-3333-3333-3333-333333333333",
		OriginalFilename: "no-smtp.mp4",
		Type:             "COMPLETED",
	})
	require.NoError(t, err)
}
