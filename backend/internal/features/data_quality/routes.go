package data_quality

import (
	"github.com/gin-gonic/gin"

	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/middleware"
)

func RegisterProtectedRoutes(
	protected *gin.RouterGroup,
	handler *Handler,
) {
	registerDQAv2Routes(protected.Group("/data-quality/dqa"), handler)
	registerDQAv2Routes(protected.Group("/data-validation/dqa"), handler)

	issues := protected.Group("/issues")
	{
		issues.POST(
			"",
			middleware.RequirePermission(authz.PermissionIssueTrackerWrite),
			handler.CreateIssue,
		)

		issues.GET(
			"",
			middleware.RequirePermission(authz.PermissionIssueTrackerRead),
			handler.ListIssues,
		)

		issues.GET(
			"/summary-by-program",
			middleware.RequirePermission(authz.PermissionIssueTrackerRead),
			handler.ListIssueSummaryByProgram,
		)

		issues.PUT(
			"/:issueCode",
			middleware.RequirePermission(authz.PermissionIssueTrackerWrite),
			handler.UpdateIssue,
		)

		issues.POST(
			"/:issueCode/resolveIssue",
			middleware.RequireAnyPermission(
				authz.PermissionIssueTrackerComment,
				authz.PermissionIssueTrackerManage,
				authz.PermissionIssueTrackerClose,
			),
			handler.ResolveIssue,
		)

		issues.GET(
			"/:issueCode/transactions",
			middleware.RequirePermission(authz.PermissionIssueTrackerRead),
			handler.ListIssueResolutionTransactions,
		)

		issues.GET(
			"/keycloak-groups",
			middleware.RequirePermission(authz.PermissionIssueTrackerRead),
			handler.ListKeycloakGroups,
		)

		issues.GET(
			"/keycloak-groups/:groupId/members",
			middleware.RequirePermission(authz.PermissionIssueTrackerRead),
			handler.ListKeycloakGroupMembers,
		)

		issues.GET(
			"/:issueCode",
			middleware.RequirePermission(authz.PermissionIssueTrackerRead),
			handler.GetIssueByCode,
		)

		issues.POST(
			"/assign",
			middleware.RequireAnyPermission(
				authz.PermissionIssueTrackerAssign,
				authz.PermissionIssueTrackerManage,
				authz.PermissionIssueTrackerWrite,
			),
			handler.AssignIssues,
		)

		issues.POST(
			"/:issueCode/assign",
			middleware.RequireAnyPermission(
				authz.PermissionIssueTrackerAssign,
				authz.PermissionIssueTrackerManage,
				authz.PermissionIssueTrackerWrite,
			),
			handler.AssignIssues,
		)
	}
}

// registerDQAv2Routes wires the ported declarative rule engine (see the
// dqa subpackage): table registration, rule CRUD + compile preview, seeding
// the built-in eCHIS check pack, and triggering/browsing scans.
func registerDQAv2Routes(group *gin.RouterGroup, handler *Handler) {
	tables := group.Group("/tables")
	{
		tables.GET("", middleware.RequirePermission(authz.PermissionDataQualityRead), handler.ListDQATables)
		tables.POST("", middleware.RequirePermission(authz.PermissionDataQualityWrite), handler.UpsertDQATable)
		tables.DELETE("/:tableId", middleware.RequirePermission(authz.PermissionDataQualityWrite), handler.DeleteDQATable)
	}

	rules := group.Group("/rules")
	{
		rules.GET("", middleware.RequirePermission(authz.PermissionDataQualityRead), handler.ListDQARules)
		rules.POST("", middleware.RequirePermission(authz.PermissionDataQualityWrite), handler.UpsertDQARule)
		rules.POST("/compile", middleware.RequirePermission(authz.PermissionDataQualityWrite), handler.CompileDQARule)
		rules.POST("/seed", middleware.RequirePermission(authz.PermissionDataQualityWrite), handler.SeedDQABuiltinRules)
		rules.DELETE("/:tableId/:code", middleware.RequirePermission(authz.PermissionDataQualityWrite), handler.DeleteDQARule)
	}

	group.POST("/run", middleware.RequirePermission(authz.PermissionDataQualityWrite), handler.RunDQATable)
	group.GET("/runs", middleware.RequirePermission(authz.PermissionDataQualityRead), handler.ListDQARuns)
	group.GET("/runs/:runId/flags", middleware.RequirePermission(authz.PermissionDataQualityRead), handler.ListDQAFlags)
	group.GET("/runs/:runId/flag-filters", middleware.RequirePermission(authz.PermissionDataQualityRead), handler.ListDQAFlagFilters)
}
