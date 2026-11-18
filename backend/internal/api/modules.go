package router

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/api/handler"
	announcementfeature "github.com/moh-sso-dashboard/internal/features/announcements"
	auditfeature "github.com/moh-sso-dashboard/internal/features/audit"
	geojsonfeature "github.com/moh-sso-dashboard/internal/features/geojson"
	metricsfeature "github.com/moh-sso-dashboard/internal/features/metrics"
	reportschedulerfeature "github.com/moh-sso-dashboard/internal/features/report_scheduler"
	"github.com/moh-sso-dashboard/internal/middleware"
	"github.com/moh-sso-dashboard/internal/ratelimit"
	"github.com/moh-sso-dashboard/module"
	announcementmodule "github.com/mohuganda/moh-sso-dashboard/backend/modules/announcements"
	auditmodule "github.com/mohuganda/moh-sso-dashboard/backend/modules/audit"
	geojsonmodule "github.com/mohuganda/moh-sso-dashboard/backend/modules/geojson"
	metricsmodule "github.com/mohuganda/moh-sso-dashboard/backend/modules/metrics"
	schedulermodule "github.com/mohuganda/moh-sso-dashboard/backend/modules/report-scheduler"
)

// NewModuleRegistry is the distribution's explicit composition root. Legacy
// features enter this inventory only as their registration is migrated.
func NewModuleRegistry(handlers HandlerSet, disabled string, limiter *ratelimit.Limiter, scheduler *reportschedulerfeature.Worker) (*module.Registry, error) {
	enabled := map[string]bool{"metrics": true, "audit": true, "geojson": true, "report-scheduler": true, "announcements": true}
	for _, id := range strings.Split(disabled, ",") {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, known := enabled[id]; !known {
			return nil, fmt.Errorf("unknown disabled module %q", id)
		}
		enabled[id] = false
	}
	return module.New(module.Definition{
		Info: module.Info{ID: "metrics", Version: metricsmodule.Version(), Enabled: enabled["metrics"]},
		Initialize: func(context.Context) (module.Stop, error) {
			if handlers.Metrics == nil {
				return nil, errors.New("metrics handler is required")
			}
			return nil, nil
		},
		Register: func(routes module.Routes) {
			metricsfeature.RegisterAdminRoutes(routes.Admin, handlers.Metrics)
		},
	}, module.Definition{
		Info: module.Info{ID: "audit", Version: auditmodule.Version(), Enabled: enabled["audit"]},
		Initialize: func(context.Context) (module.Stop, error) {
			if handlers.Audit == nil {
				return nil, errors.New("audit read handler is required")
			}
			return nil, nil
		},
		Register: func(routes module.Routes) {
			auditfeature.RegisterAdminRoutes(routes.Admin, handlers.Audit, limiter)
		},
	}, module.Definition{
		Info: module.Info{ID: "geojson", Version: geojsonmodule.Version(), Enabled: enabled["geojson"]},
		Initialize: func(context.Context) (module.Stop, error) {
			if handlers.GeoJSON == nil {
				return nil, errors.New("geojson handler is required")
			}
			return nil, nil
		},
		Register: func(routes module.Routes) {
			geojsonfeature.RegisterProtectedRoutes(routes.Protected, handlers.GeoJSON)
		},
	}, module.Definition{
		Info: module.Info{ID: "report-scheduler", Version: schedulermodule.Version(), Enabled: enabled["report-scheduler"]},
		Initialize: func(ctx context.Context) (module.Stop, error) {
			if handlers.ReportScheduler == nil {
				return nil, errors.New("report scheduler handler is required")
			}
			if scheduler == nil {
				return nil, errors.New("report scheduler worker is required")
			}
			return scheduler.Initialize(ctx)
		},
		Register: func(routes module.Routes) {
			if routes.Protected != nil {
				reportschedulerfeature.RegisterProtectedRoutes(routes.Protected, handlers.ReportScheduler)
			}
			if routes.Public != nil {
				reportschedulerfeature.RegisterPublicRoutes(routes.Public, handlers.ReportScheduler, limiter)
			}
		},
	}, module.Definition{
		Info: module.Info{ID: "announcements", Version: announcementmodule.Version(), Enabled: enabled["announcements"]},
		Initialize: func(context.Context) (module.Stop, error) {
			if handlers.Announcements == nil {
				return nil, errors.New("announcements handler is required")
			}
			return nil, nil
		},
		Register: func(routes module.Routes) {
			if routes.Public != nil {
				announcementfeature.RegisterPublicRoutes(routes.Public, handlers.Announcements)
			}
			if routes.Protected != nil {
				announcementfeature.RegisterProtectedRoutes(routes.Protected, handlers.Announcements)
			}
			if routes.Admin != nil {
				announcementfeature.RegisterAdminRoutes(routes.Admin, handlers.Announcements, limiter)
			}
		},
	})
}

func registerModuleInventory(protected *gin.RouterGroup, registry *module.Registry) {
	h := &handler.ModuleHandler{Registry: registry}
	protected.GET("/admin/modules", middleware.RequireAdmin(), h.List)
}
