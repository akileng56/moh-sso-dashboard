package routes

import (
	"github.com/gin-gonic/gin"

	adminunitsfeature "github.com/moh-sso-dashboard/internal/adapters/modules/admin_units"
	"github.com/moh-sso-dashboard/internal/authz"
	visualiserfeature "github.com/moh-sso-dashboard/internal/features/visualiser"
	"github.com/moh-sso-dashboard/internal/middleware"
)

// VisualiserGroup retains the common read permissions for visualizer consumers.
func VisualiserGroup(protected *gin.RouterGroup) *gin.RouterGroup {
	visualiser := protected.Group("/visualizer")
	visualiser.Use(middleware.RequireAnyPermission(
		authz.PermissionReportBrowserRead,
		authz.PermissionDocumentsRead,
		authz.PermissionDataQualityRead,
		authz.PermissionSurveillanceRead,
	))
	return visualiser
}

func registerVisualiserRoutes(protected *gin.RouterGroup, deps Dependencies) {
	visualiser := VisualiserGroup(protected)
	if deps.Modules == nil {
		adminunitsfeature.RegisterVisualiserRoutes(visualiser, deps.AdminUnits)
	}
	visualiserfeature.RegisterRoutes(visualiser, deps.Visualiser)
}
