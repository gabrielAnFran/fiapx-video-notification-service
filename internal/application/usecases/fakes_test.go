package usecases

import (
	"context"
	"sync"

	"github.com/gabrielAnFran/fiapx-video-notification-service/internal/domain/entities"
	"github.com/gabrielAnFran/fiapx-video-notification-service/internal/infrastructure/email"
	"github.com/google/uuid"
)

// fakeNotificationRepository is an in-memory repositories.NotificationRepository.
type fakeNotificationRepository struct {
	mu            sync.Mutex
	notifications []entities.Notification
}

func newFakeNotificationRepository() *fakeNotificationRepository {
	return &fakeNotificationRepository{}
}

func (f *fakeNotificationRepository) Create(_ context.Context, n *entities.Notification) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.notifications = append(f.notifications, *n)
	return nil
}

// fakeProcessedEventRepository is an in-memory repositories.ProcessedEventRepository.
type fakeProcessedEventRepository struct {
	mu        sync.Mutex
	processed map[uuid.UUID]bool
}

func newFakeProcessedEventRepository() *fakeProcessedEventRepository {
	return &fakeProcessedEventRepository{processed: map[uuid.UUID]bool{}}
}

func (f *fakeProcessedEventRepository) IsProcessed(_ context.Context, eventID uuid.UUID) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.processed[eventID], nil
}

func (f *fakeProcessedEventRepository) MarkProcessed(_ context.Context, eventID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.processed[eventID] = true
	return nil
}

// fakeSender is an in-memory EmailSender.
type fakeSender struct {
	mu      sync.Mutex
	sent    []email.Notification
	sendErr error
}

func newFakeSender() *fakeSender {
	return &fakeSender{}
}

func (f *fakeSender) Send(_ string, n email.Notification) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, n)
	return f.sendErr
}
