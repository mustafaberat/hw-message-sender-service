package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"

	"message-sender/config"
	"message-sender/model"
)

type Repository struct {
	client             *redis.Client
	serviceStatusKey   string
	sentMessagesPrefix string
	messageCacheTTL    time.Duration
}

func NewRepository(client *redis.Client, cfg *config.RedisConfig) *Repository {
	return &Repository{
		client:             client,
		serviceStatusKey:   cfg.ServiceStatusKey,
		sentMessagesPrefix: cfg.SentMessagesPrefix,
		messageCacheTTL:    cfg.MessageCacheTTL,
	}
}

func (r *Repository) GetServiceStatus(ctx context.Context) (model.ServiceStatus, error) {
	status, err := r.client.Get(ctx, r.serviceStatusKey).Result()
	if errors.Is(err, redis.Nil) {
		return model.StatusStopped, nil
	}
	if err != nil {
		return "", err
	}
	return model.ServiceStatus(status), nil
}

func (r *Repository) SetServiceStatus(ctx context.Context, status model.ServiceStatus) error {
	return r.client.Set(ctx, r.serviceStatusKey, string(status), 0).Err()
}

func (r *Repository) CacheMessageSent(ctx context.Context, messageID string, sentAt time.Time) error {
	key := r.sentMessagesPrefix + messageID
	data, err := json.Marshal(map[string]interface{}{
		"messageId": messageID,
		"sentAt":    sentAt,
	})
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, r.messageCacheTTL).Err()
}

func (r *Repository) IsMessageSent(ctx context.Context, messageID string) (bool, error) {
	key := r.sentMessagesPrefix + messageID
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (r *Repository) GetCachedSentMessages(ctx context.Context) (map[string]time.Time, error) {
	keys, err := r.client.Keys(ctx, r.sentMessagesPrefix+"*").Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string]time.Time)
	for _, key := range keys {
		data, err := r.client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var messageData map[string]interface{}
		if err := json.Unmarshal([]byte(data), &messageData); err != nil {
			continue
		}

		messageID, ok := messageData["messageId"].(string)
		if !ok {
			continue
		}

		sentAtStr, ok := messageData["sentAt"].(string)
		if !ok {
			continue
		}

		sentAt, err := time.Parse(time.RFC3339, sentAtStr)
		if err != nil {
			continue
		}

		result[messageID] = sentAt
	}

	return result, nil
}
