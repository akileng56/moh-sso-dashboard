package cache

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"

	"github.com/moh-sso-dashboard/internal/model"
	"github.com/moh-sso-dashboard/internal/ws"
)

type NotificationSubscriber struct {
	redis *redis.Client
	hub   *ws.Hub
}

func NewNotificationSubscriber(
	redis *redis.Client,
	hub *ws.Hub,
) *NotificationSubscriber {
	return &NotificationSubscriber{
		redis: redis,
		hub:   hub,
	}
}

func (s *NotificationSubscriber) Start(ctx context.Context) {

	pubsub := s.redis.Subscribe(
		ctx,
		"notify:global",
		"notify:role:*",
		"notify:user:*",
		"notify:client:*",
	)

	go func() {
		for msg := range pubsub.Channel() {
			payload := []byte(msg.Payload)

			var n model.Notification
			if err := json.Unmarshal([]byte(payload), &n); err != nil {
				log.Printf("invalid notification payload: %v", err)
				continue
			}

			s.dispatch(n, payload)
		}
	}()
}

func (s *NotificationSubscriber) dispatch(
	n model.Notification,
	payload []byte,
) {
	s.hub.Broadcast(func(c *ws.Client) bool {

		switch {
		case n.UserID != "":
			return c.UserID == n.UserID

		case n.TargetRole != "":
			return c.Role == n.TargetRole

		case n.ClientID != "":
			return c.ClientID == n.ClientID

		default:
			return true
		}

	}, payload)
}
