package cache

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"

	"github.com/moh-sso-dashboard/internal/model"
)

type NotificationPublisher struct {
	redis *redis.Client
}

func NewNotificationPublisher(redis *redis.Client) *NotificationPublisher {
	return &NotificationPublisher{redis: redis}
}

func (p *NotificationPublisher) Publish(
	ctx context.Context,
	channel string,
	notification model.Notification,
) error {

	payload, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	return p.redis.Publish(ctx, channel, payload).Err()
}
