package service

import (
	"context"

	"message-sender/model"
)

type Service interface {
	StartService(ctx context.Context) error
	StopService(ctx context.Context) error
	GetServiceStatus(ctx context.Context) (model.ServiceStatus, error)
	GetSentMessages(ctx context.Context, page, limit int) (*model.SentMessagesResponse, error)
}
