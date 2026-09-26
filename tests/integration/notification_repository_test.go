//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/gabrielAnFran/fiapx-video-notification-service/internal/domain/entities"
	infradb "github.com/gabrielAnFran/fiapx-video-notification-service/internal/infrastructure/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationRepository_Create_PersistsRow(t *testing.T) {
	db := newTestDB(t)
	truncateAll(t, db)
	repo := infradb.NewNotificationRepository(db)
	ctx := context.Background()

	n := &entities.Notification{
		ID:        uuid.New(),
		VideoID:   uuid.New(),
		UserID:    uuid.New(),
		Recipient: "someone@example.com",
		Channel:   "email",
		Type:      "COMPLETED",
		Status:    "SENT",
		SentAt:    time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	}
	require.NoError(t, repo.Create(ctx, n))

	// NotificationRepository exposes no read method, so assert persistence
	// directly against the notification_log table (as an audit record would
	// be inspected).
	var count int64
	require.NoError(t, db.Table("notification_log").Where("id = ?", n.ID).Count(&count).Error)
	assert.Equal(t, int64(1), count)

	var got struct {
		Recipient string
		Channel   string
		Type      string
		Status    string
	}
	require.NoError(t, db.Table("notification_log").
		Select("recipient, channel, type, status").
		Where("id = ?", n.ID).
		Scan(&got).Error)
	assert.Equal(t, "someone@example.com", got.Recipient)
	assert.Equal(t, "email", got.Channel)
	assert.Equal(t, "COMPLETED", got.Type)
	assert.Equal(t, "SENT", got.Status)
}

func TestNotificationRepository_Create_FailedWithErrorMessage(t *testing.T) {
	db := newTestDB(t)
	truncateAll(t, db)
	repo := infradb.NewNotificationRepository(db)
	ctx := context.Background()

	errMsg := "smtp: connection refused"
	n := &entities.Notification{
		ID:           uuid.New(),
		VideoID:      uuid.New(),
		UserID:       uuid.New(),
		Recipient:    "fail@example.com",
		Channel:      "email",
		Type:         "FAILED",
		Status:       "FAILED",
		ErrorMessage: &errMsg,
		SentAt:       time.Now().UTC(),
		CreatedAt:    time.Now().UTC(),
	}
	require.NoError(t, repo.Create(ctx, n))

	var got struct {
		Status       string
		ErrorMessage *string
	}
	require.NoError(t, db.Table("notification_log").
		Select("status, error_message").
		Where("id = ?", n.ID).
		Scan(&got).Error)
	assert.Equal(t, "FAILED", got.Status)
	require.NotNil(t, got.ErrorMessage)
	assert.Equal(t, errMsg, *got.ErrorMessage)
}

func TestProcessedEventRepository_IsProcessed_MarkProcessed_Idempotent(t *testing.T) {
	db := newTestDB(t)
	truncateAll(t, db)
	repo := infradb.NewNotificationRepository(db)
	ctx := context.Background()

	eventID := uuid.New()

	processed, err := repo.IsProcessed(ctx, eventID)
	require.NoError(t, err)
	assert.False(t, processed)

	require.NoError(t, repo.MarkProcessed(ctx, eventID))
	require.NoError(t, repo.MarkProcessed(ctx, eventID)) // must be idempotent on conflict, not error

	processed, err = repo.IsProcessed(ctx, eventID)
	require.NoError(t, err)
	assert.True(t, processed)
}

func TestProcessedEventRepository_IsProcessed_UnknownEventIsFalse(t *testing.T) {
	db := newTestDB(t)
	truncateAll(t, db)
	repo := infradb.NewNotificationRepository(db)

	processed, err := repo.IsProcessed(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.False(t, processed)
}
