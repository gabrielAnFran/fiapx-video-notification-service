package usecases

import (
	"context"
	"errors"
	"testing"

	"github.com/gabrielAnFran/fiapx-video-notification-service/internal/infrastructure/messaging"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newNotifyCommand(t *testing.T, videoID, userID uuid.UUID, notifType string) messaging.Event {
	t.Helper()
	ev, err := messaging.NewEvent(EventNotifyRequested, "corr-x", "", notifyRequestedPayload{
		VideoID:          videoID.String(),
		UserID:           userID.String(),
		RecipientEmail:   "user@example.com",
		NotificationType: notifType,
		OriginalFilename: "movie.mp4",
	})
	require.NoError(t, err)
	return ev
}

func TestSendNotification_Success(t *testing.T) {
	notifications := newFakeNotificationRepository()
	processed := newFakeProcessedEventRepository()
	sender := newFakeSender()
	uc := NewSendNotificationUseCase(notifications, processed, sender)

	videoID, userID := uuid.New(), uuid.New()
	ev := newNotifyCommand(t, videoID, userID, "COMPLETED")

	require.NoError(t, uc.Handle(context.Background(), ev))

	require.Len(t, notifications.notifications, 1)
	assert.Equal(t, "SENT", notifications.notifications[0].Status)
	assert.Nil(t, notifications.notifications[0].ErrorMessage)
	assert.Equal(t, "email", notifications.notifications[0].Channel)

	require.Len(t, sender.sent, 1)

	eventID, err := uuid.Parse(ev.EventID)
	require.NoError(t, err)
	isProcessed, err := processed.IsProcessed(context.Background(), eventID)
	require.NoError(t, err)
	assert.True(t, isProcessed)
}

func TestSendNotification_SendFailure(t *testing.T) {
	notifications := newFakeNotificationRepository()
	processed := newFakeProcessedEventRepository()
	sender := newFakeSender()
	sender.sendErr = errors.New("smtp down")
	uc := NewSendNotificationUseCase(notifications, processed, sender)

	videoID, userID := uuid.New(), uuid.New()
	ev := newNotifyCommand(t, videoID, userID, "FAILED")

	err := uc.Handle(context.Background(), ev)
	require.Error(t, err)

	require.Len(t, notifications.notifications, 1)
	assert.Equal(t, "FAILED", notifications.notifications[0].Status)
	require.NotNil(t, notifications.notifications[0].ErrorMessage)
	assert.Equal(t, "smtp down", *notifications.notifications[0].ErrorMessage)

	eventID, perr := uuid.Parse(ev.EventID)
	require.NoError(t, perr)
	isProcessed, err := processed.IsProcessed(context.Background(), eventID)
	require.NoError(t, err)
	assert.False(t, isProcessed, "must not be marked processed on send failure, so amqp retry/DLQ can redeliver")
}

func TestSendNotification_AlreadyProcessed(t *testing.T) {
	notifications := newFakeNotificationRepository()
	processed := newFakeProcessedEventRepository()
	sender := newFakeSender()
	uc := NewSendNotificationUseCase(notifications, processed, sender)

	videoID, userID := uuid.New(), uuid.New()
	ev := newNotifyCommand(t, videoID, userID, "COMPLETED")
	eventID, err := uuid.Parse(ev.EventID)
	require.NoError(t, err)
	require.NoError(t, processed.MarkProcessed(context.Background(), eventID))

	require.NoError(t, uc.Handle(context.Background(), ev))

	assert.Empty(t, notifications.notifications, "should short-circuit before writing a log row")
	assert.Empty(t, sender.sent, "should short-circuit before attempting to send")
}
