package db

import (
	"context"
	"time"

	"github.com/gabrielAnFran/fiapx-video-notification-service/internal/domain/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type notificationLogModel struct {
	ID           uuid.UUID `gorm:"column:id;primaryKey"`
	VideoID      uuid.UUID `gorm:"column:video_id"`
	UserID       uuid.UUID `gorm:"column:user_id"`
	Recipient    string    `gorm:"column:recipient"`
	Channel      string    `gorm:"column:channel"`
	Type         string    `gorm:"column:type"`
	Status       string    `gorm:"column:status"`
	ErrorMessage *string   `gorm:"column:error_message"`
	SentAt       time.Time `gorm:"column:sent_at"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}

func (notificationLogModel) TableName() string { return "notification_log" }

type processedEventModel struct {
	EventID     uuid.UUID `gorm:"column:event_id;primaryKey"`
	ProcessedAt time.Time `gorm:"column:processed_at"`
}

func (processedEventModel) TableName() string { return "processed_events" }

// NotificationRepository is a GORM-backed implementation of both
// repositories.NotificationRepository and repositories.ProcessedEventRepository,
// mirroring the pattern used by pos-os-service's OrderRepository of a single
// type implementing several small interfaces.
type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *entities.Notification) error {
	m := notificationLogModel{
		ID:           n.ID,
		VideoID:      n.VideoID,
		UserID:       n.UserID,
		Recipient:    n.Recipient,
		Channel:      n.Channel,
		Type:         n.Type,
		Status:       n.Status,
		ErrorMessage: n.ErrorMessage,
		SentAt:       n.SentAt,
		CreatedAt:    n.CreatedAt,
	}
	return r.db.WithContext(ctx).Create(&m).Error
}

// ProcessedEventRepository implementation.

func (r *NotificationRepository) IsProcessed(ctx context.Context, eventID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&processedEventModel{}).Where("event_id = ?", eventID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *NotificationRepository) MarkProcessed(ctx context.Context, eventID uuid.UUID) error {
	m := processedEventModel{EventID: eventID, ProcessedAt: time.Now().UTC()}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&m).Error
}
