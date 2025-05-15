package repository

import (
	"context"
	"time"

	"message-sender/model"
)

type Repository interface {
	GetUnsentMessages(ctx context.Context, limit int) ([]model.Message, error)
	MarkMessageAsSent(ctx context.Context, id uint, messageID string, sentAt time.Time) error
	GetSentMessages(ctx context.Context, page, limit int) ([]model.Message, int, error)
	GetMessageByID(ctx context.Context, id uint) (*model.Message, error)
	SaveMessage(ctx context.Context, message *model.Message) error
}

type ServiceStatusRepository interface {
	GetServiceStatus(ctx context.Context) (model.ServiceStatus, error)
	SetServiceStatus(ctx context.Context, status model.ServiceStatus) error
}

type CacheRepository interface {
	CacheMessageSent(ctx context.Context, messageID string, sentAt time.Time) error
	IsMessageSent(ctx context.Context, messageID string) (bool, error)
	GetCachedSentMessages(ctx context.Context) (map[string]time.Time, error)
}
