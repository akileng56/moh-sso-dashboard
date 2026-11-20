package routes

import (
	"github.com/gin-gonic/gin"

	reportschedulerfeature "github.com/moh-sso-dashboard/internal/adapters/modules/report_scheduler"
)

func RegisterPublicReportSchedulerRoutes(api *gin.RouterGroup, deps Dependencies) {
	reportschedulerfeature.RegisterPublicRoutes(api, deps.ReportScheduler, deps.Limiter)
}
