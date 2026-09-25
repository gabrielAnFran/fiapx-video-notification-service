package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gabrielAnFran/fiapx-video-notification-service/internal/domain/entities"
	"github.com/gabrielAnFran/fiapx-video-notification-service/internal/domain/repositories"
	"github.com/gabrielAnFran/fiapx-video-notification-service/internal/infrastructure/email"
	"github.com/gabrielAnFran/fiapx-video-notification-service/internal/infrastructure/messaging"
	"github.com/google/uuid"
)

// EventNotifyRequested is the routing key / event name this use case
// consumes; also used by cmd/worker to declare the queue binding and route
// incoming events.
const EventNotifyRequested = "video.notify.requested"

// EmailSender is the minimal surface SendNotificationUseCase needs from
// *email.Sender. Extracted as an interface so unit tests can fake it without
// a real SMTP connection; *email.Sender satisfies it as-is.
type EmailSender interface {
	Send(to string, n email.Notification) error
}

type SendNotificationUseCase struct {
	notifications   repositories.NotificationRepository
	processedEvents repositories.ProcessedEventRepository
	sender          EmailSender
}

func NewSendNotificationUseCase(notifications repositories.NotificationRepository, processedEvents repositories.ProcessedEventRepository, sender EmailSender) *SendNotificationUseCase {
	return &SendNotificationUseCase{notifications: notifications, processedEvents: processedEvents, sender: sender}
}

type notifyRequestedPayload struct {
	VideoID          string `json:"video_id"`
	UserID           string `json:"user_id"`
	RecipientEmail   string `json:"recipient_email"`
	NotificationType string `json:"notification_type"` // "COMPLETED" or "FAILED"
	OriginalFilename string `json:"original_filename"`
	ErrorMessage     string `json:"error_message,omitempty"`
}

// Handle processes a video.notify.requested command, idempotently: it checks
// processed_events before doing any work.
//
// It marks the event as processed only *after* a successful send. This
// matters: if MarkProcessed happened before (or regardless of) the send
// outcome, a transient SMTP failure would never be retried, because the
// idempotency check would short-circuit it on redelivery. Instead, on send
// failure we still record a notification_log row (Status "FAILED") for
// audit purposes, then return the error so amqp.go's retry/DLQ mechanism
// handles redelivery. After MaxRetries the message lands in the DLQ, which
// is the correct terminal behavior here — the user's video status is
// already durably COMPLETED/FAILED in upload-service regardless of whether
// the notification email made it, so this retry/DLQ path only affects the
// "nice to have" notification, never the correctness of the core system.
func (uc *SendNotificationUseCase) Handle(ctx context.Context, ev messaging.Event) error {
	eventID, err := uuid.Parse(ev.EventID)
	if err != nil {
		return fmt.Errorf("invalid event id: %w", err)
	}

	processed, err := uc.processedEvents.IsProcessed(ctx, eventID)
	if err != nil {
		return fmt.Errorf("check processed: %w", err)
	}
	if processed {
		return nil
	}

	var cmd notifyRequestedPayload
	if err := json.Unmarshal(ev.Payload, &cmd); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}
	videoID, err := uuid.Parse(cmd.VideoID)
	if err != nil {
		return fmt.Errorf("invalid video_id: %w", err)
	}
	userID, err := uuid.Parse(cmd.UserID)
	if err != nil {
		return fmt.Errorf("invalid user_id: %w", err)
	}

	sendErr := uc.sender.Send(cmd.RecipientEmail, email.Notification{
		VideoID:          cmd.VideoID,
		OriginalFilename: cmd.OriginalFilename,
		Type:             cmd.NotificationType,
		ErrorMessage:     cmd.ErrorMessage,
	})

	now := time.Now().UTC()
	notification := &entities.Notification{
		ID:        uuid.New(),
		VideoID:   videoID,
		UserID:    userID,
		Recipient: cmd.RecipientEmail,
		Channel:   "email",
		Type:      cmd.NotificationType,
		SentAt:    now,
		CreatedAt: now,
	}
	if sendErr != nil {
		msg := sendErr.Error()
		notification.Status = "FAILED"
		notification.ErrorMessage = &msg
	} else {
		notification.Status = "SENT"
	}

	if err := uc.notifications.Create(ctx, notification); err != nil {
		return fmt.Errorf("save notification log: %w", err)
	}

	if sendErr != nil {
		return fmt.Errorf("send email: %w", sendErr)
	}

	if err := uc.processedEvents.MarkProcessed(ctx, eventID); err != nil {
		return fmt.Errorf("mark processed: %w", err)
	}
	return nil
}
