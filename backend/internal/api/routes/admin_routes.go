package routes

import (
	"github.com/gin-gonic/gin"

	clientfeature "github.com/moh-sso-dashboard/internal/features/clients"
	healthcontextfeature "github.com/moh-sso-dashboard/internal/features/health_context"
	notificationsfeature "github.com/moh-sso-dashboard/internal/features/notifications"
	rbacfeature "github.com/moh-sso-dashboard/internal/features/rbac"
	userfeature "github.com/moh-sso-dashboard/internal/features/users"
	"github.com/moh-sso-dashboard/internal/modulehttp"
)

func RegisterAdminRoutes(protected *gin.RouterGroup, deps Dependencies) {
	admin := protected.Group("/admin")

	registerAdminUserRoutes(admin, deps)
	registerAdminClientRoleRoutes(admin, deps)
	if deps.Modules == nil {
		registerAdminMetricRoutes(protected, deps)
		registerAdminAuditRoutes(protected, deps)
	}
	registerAdminNotificationRoutes(admin, deps)

	registerAdminRBACRoutes(admin, deps)
	registerAdminHealthContextRoutes(admin, deps)
}

func registerAdminHealthContextRoutes(admin *gin.RouterGroup, deps Dependencies) {
	healthcontextfeature.RegisterAdminRoutes(admin, deps.HealthContexts)
}

func registerAdminUserRoutes(admin *gin.RouterGroup, deps Dependencies) {
	userfeature.RegisterAdminRoutes(admin, deps.Users, deps.Limiter)
}

func registerAdminClientRoleRoutes(admin *gin.RouterGroup, deps Dependencies) {
	clientfeature.RegisterAdminRoutes(admin, deps.Clients)
}

func registerAdminMetricRoutes(protected *gin.RouterGroup, deps Dependencies) {
	if deps.Metrics != nil {
		modulehttp.Register(protected, deps.Metrics)
	}
}

func registerAdminAuditRoutes(protected *gin.RouterGroup, deps Dependencies) {
	if deps.Audit != nil {
		modulehttp.Register(protected, deps.Audit)
	}
}

func registerAdminNotificationRoutes(admin *gin.RouterGroup, deps Dependencies) {
	notificationsfeature.RegisterAdminRoutes(admin, deps.Notifications, deps.Limiter)
}

func registerAdminRBACRoutes(admin *gin.RouterGroup, deps Dependencies) {
	rbacfeature.RegisterAdminRoutes(admin, deps.RBAC, deps.Limiter)
}
