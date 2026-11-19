package routes

import (
	"github.com/gin-gonic/gin"

	announcementfeature "github.com/moh-sso-dashboard/internal/adapters/modules/announcements"
	reportschedulerfeature "github.com/moh-sso-dashboard/internal/adapters/modules/report_scheduler"
)

func RegisterPublicAnnouncementRoutes(api *gin.RouterGroup, deps Dependencies) {
	if deps.Modules == nil {
		announcementfeature.RegisterPublicRoutes(api, deps.Announcements)
	}
}

func RegisterPublicReportSchedulerRoutes(api *gin.RouterGroup, deps Dependencies) {
	reportschedulerfeature.RegisterPublicRoutes(api, deps.ReportScheduler, deps.Limiter)
}
