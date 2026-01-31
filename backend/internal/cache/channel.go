package cache

import "github.com/moh-sso-dashboard/internal/model"

func ResolveNotificationChannel(n model.Notification) string {
	switch {
	case n.UserID != "":
		return "notify:user:" + n.UserID

	case n.TargetRole != "":
		return "notify:role:" + n.TargetRole

	case n.ClientID != "":
		return "notify:client:" + n.ClientID

	default:
		return "notify:global"
	}
}
