package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"message-sender/config"
	"message-sender/model"
	"message-sender/repository"
)

type MessageProcessor struct {
	repo          repository.Repository
	statusRepo    repository.ServiceStatusRepository
	cacheRepo     repository.CacheRepository
	logger        *zap.Logger
	cfg           *config.Config
	ticker        *time.Ticker
	stopChan      chan struct{}
	processingMux sync.Mutex
	httpClient    *http.Client
}

func NewMessageProcessor(
	repo repository.Repository,
	statusRepo repository.ServiceStatusRepository,
	cacheRepo repository.CacheRepository,
	logger *zap.Logger,
	cfg *config.Config,
) *MessageProcessor {
	return &MessageProcessor{
		repo:       repo,
		statusRepo: statusRepo,
		cacheRepo:  cacheRepo,
		logger:     logger,
		cfg:        cfg,
		stopChan:   make(chan struct{}),
		httpClient: &http.Client{
			Timeout: cfg.Webhook.Timeout,
		},
	}
}

func (s *MessageProcessor) StartService(ctx context.Context) error {
	s.processingMux.Lock()
	defer s.processingMux.Unlock()

	status, err := s.statusRepo.GetServiceStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get service status: %w", err)
	}

	if status == model.StatusRunning {
		return nil
	}

	if err := s.statusRepo.SetServiceStatus(ctx, model.StatusRunning); err != nil {
		return fmt.Errorf("failed to set service status: %w", err)
	}

	s.ticker = time.NewTicker(s.cfg.Message.ProcessInterval)

	go func() {
		s.processMessages(context.Background())

		for {
			select {
			case <-s.ticker.C:
				s.processMessages(context.Background())
			case <-s.stopChan:
				s.ticker.Stop()
				return
			}
		}
	}()

	s.logger.Info("Message sending service started")
	return nil
}

func (s *MessageProcessor) StopService(ctx context.Context) error {
	s.processingMux.Lock()
	defer s.processingMux.Unlock()

	status, err := s.statusRepo.GetServiceStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get service status: %w", err)
	}

	if status == model.StatusStopped {
		return nil
	}

	if s.ticker != nil {
		close(s.stopChan)
	}

	if err := s.statusRepo.SetServiceStatus(ctx, model.StatusStopped); err != nil {
		return fmt.Errorf("failed to set service status: %w", err)
	}

	s.logger.Info("Message sending service stopped")
	return nil
}

func (s *MessageProcessor) GetServiceStatus(ctx context.Context) (model.ServiceStatus, error) {
	return s.statusRepo.GetServiceStatus(ctx)
}

func (s *MessageProcessor) GetSentMessages(ctx context.Context, page, limit int) (*model.SentMessagesResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	messages, total, err := s.repo.GetSentMessages(ctx, page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get sent messages: %w", err)
	}

	return &model.SentMessagesResponse{
		Messages: messages,
		Count:    total,
	}, nil
}

func (s *MessageProcessor) processMessages(ctx context.Context) {
	s.logger.Debug("Processing messages")

	messages, err := s.repo.GetUnsentMessages(ctx, s.cfg.Message.BatchSize)
	if err != nil {
		s.logger.Error("Failed to get unsent messages", zap.Error(err))
		return
	}

	if len(messages) == 0 {
		s.logger.Debug("No unsent messages found")
		return
	}

	s.logger.Debug("Found unsent messages", zap.Int("count", len(messages)))

	for _, msg := range messages {
		if msg.MessageID != "" {
			sent, err := s.cacheRepo.IsMessageSent(ctx, msg.MessageID)
			if err != nil {
				s.logger.Error("Failed to check if message is sent", zap.Error(err), zap.Uint("messageID", msg.ID))
				continue
			}
			if sent {
				s.logger.Debug("Message already sent according to cache", zap.Uint("messageID", msg.ID))
				continue
			}
		}

		messageID, err := s.sendMessage(ctx, msg)
		if err != nil {
			s.logger.Error("Failed to send message", zap.Error(err), zap.Uint("messageID", msg.ID))
			continue
		}

		sentAt := time.Now()
		if err := s.repo.MarkMessageAsSent(ctx, msg.ID, messageID, sentAt); err != nil {
			s.logger.Error("Failed to mark message as sent", zap.Error(err), zap.Uint("messageID", msg.ID))
			continue
		}

		if err := s.cacheRepo.CacheMessageSent(ctx, messageID, sentAt); err != nil {
			s.logger.Error("Failed to cache sent message", zap.Error(err), zap.String("messageID", messageID))
		} else {
			s.logger.Info("Message successfully cached",
				zap.String("messageID", messageID),
				zap.Time("sentAt", sentAt),
				zap.Uint("msgID", msg.ID))
		}

		s.logger.Info("Message sent successfully", zap.Uint("messageID", msg.ID), zap.String("externalID", messageID))
	}
}

func (s *MessageProcessor) sendMessage(ctx context.Context, msg model.Message) (string, error) {
	payload := map[string]interface{}{
		"content":   msg.Content,
		"recipient": msg.Recipient,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.Webhook.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send webhook request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("webhook returned non-success status: %d", resp.StatusCode)
	}

	requestID := resp.Header.Get("X-Request-Id")
	if requestID != "" {
		return requestID, nil
	}

	uniqueID := fmt.Sprintf("webhook-%d-%d", msg.ID, time.Now().UnixNano())
	return uniqueID, nil
}
