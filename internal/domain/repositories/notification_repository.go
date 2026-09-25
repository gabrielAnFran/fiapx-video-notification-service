package repositories

import (
	"context"

	"github.com/gabrielAnFran/fiapx-video-notification-service/internal/domain/entities"
	"github.com/google/uuid"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *entities.Notification) error
}

type ProcessedEventRepository interface {
	IsProcessed(ctx context.Context, eventID uuid.UUID) (bool, error)
	MarkProcessed(ctx context.Context, eventID uuid.UUID) error
}
