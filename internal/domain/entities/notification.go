package entities

import (
	"time"

	"github.com/google/uuid"
)

// Notification is an audit record of a single notification attempt (e.g. an
// email sent, or attempted, to a user about a video's processing outcome).
type Notification struct {
	ID           uuid.UUID
	VideoID      uuid.UUID
	UserID       uuid.UUID
	Recipient    string
	Channel      string // e.g. "email"
	Type         string // "COMPLETED" or "FAILED"
	Status       string // "SENT" or "FAILED"
	ErrorMessage *string
	SentAt       time.Time
	CreatedAt    time.Time
}
