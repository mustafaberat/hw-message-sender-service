package service

import (
	"context"
	"testing"
	"time"

	"message-sender/config"
	"message-sender/model"

	"go.uber.org/zap/zaptest"
)

type MockRepository struct {
	messages         []model.Message
	markAsSentCalled bool
	messageID        string
}

func (m *MockRepository) GetUnsentMessages(ctx context.Context, limit int) ([]model.Message, error) {
	return m.messages, nil
}

func (m *MockRepository) MarkMessageAsSent(ctx context.Context, id uint, messageID string, sentAt time.Time) error {
	m.markAsSentCalled = true
	m.messageID = messageID
	return nil
}

func (m *MockRepository) GetSentMessages(ctx context.Context, page, limit int) ([]model.Message, int, error) {
	return []model.Message{}, 0, nil
}

func (m *MockRepository) GetMessageByID(ctx context.Context, id uint) (*model.Message, error) {
	return nil, nil
}

func (m *MockRepository) SaveMessage(ctx context.Context, message *model.Message) error {
	return nil
}

type MockStatusRepository struct {
	status model.ServiceStatus
}

func (m *MockStatusRepository) GetServiceStatus(ctx context.Context) (model.ServiceStatus, error) {
	return m.status, nil
}

func (m *MockStatusRepository) SetServiceStatus(ctx context.Context, status model.ServiceStatus) error {
	m.status = status
	return nil
}

type MockCacheRepository struct {
	cachedMessages map[string]time.Time
}

func (m *MockCacheRepository) CacheMessageSent(ctx context.Context, messageID string, sentAt time.Time) error {
	if m.cachedMessages == nil {
		m.cachedMessages = make(map[string]time.Time)
	}
	m.cachedMessages[messageID] = sentAt
	return nil
}

func (m *MockCacheRepository) IsMessageSent(ctx context.Context, messageID string) (bool, error) {
	if m.cachedMessages == nil {
		return false, nil
	}
	_, exists := m.cachedMessages[messageID]
	return exists, nil
}

func (m *MockCacheRepository) GetCachedSentMessages(ctx context.Context) (map[string]time.Time, error) {
	return m.cachedMessages, nil
}

func TestMessageProcessor_GetServiceStatus(t *testing.T) {
	mockRepo := &MockRepository{}
	mockStatusRepo := &MockStatusRepository{status: model.StatusRunning}
	mockCacheRepo := &MockCacheRepository{}
	logger := zaptest.NewLogger(t)
	cfg := &config.Config{}

	processor := NewMessageProcessor(mockRepo, mockStatusRepo, mockCacheRepo, logger, cfg)

	status, err := processor.GetServiceStatus(context.Background())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if status != model.StatusRunning {
		t.Errorf("Expected status %s, got %s", model.StatusRunning, status)
	}
}

func TestMessageProcessor_StartStopService(t *testing.T) {
	mockRepo := &MockRepository{}
	mockStatusRepo := &MockStatusRepository{status: model.StatusStopped}
	mockCacheRepo := &MockCacheRepository{}
	logger := zaptest.NewLogger(t)

	cfg := &config.Config{
		Message: config.MessageConfig{
			ProcessInterval: 100 * time.Millisecond,
		},
	}

	processor := NewMessageProcessor(mockRepo, mockStatusRepo, mockCacheRepo, logger, cfg)

	err := processor.StartService(context.Background())
	if err != nil {
		t.Fatalf("Expected no error on start, got %v", err)
	}

	status, _ := processor.GetServiceStatus(context.Background())
	if status != model.StatusRunning {
		t.Errorf("Expected status %s after start, got %s", model.StatusRunning, status)
	}

	err = processor.StopService(context.Background())
	if err != nil {
		t.Fatalf("Expected no error on stop, got %v", err)
	}

	status, _ = processor.GetServiceStatus(context.Background())
	if status != model.StatusStopped {
		t.Errorf("Expected status %s after stop, got %s", model.StatusStopped, status)
	}
}
