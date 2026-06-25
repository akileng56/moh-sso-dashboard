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
	issues := protected.Group("/issues")
	{
		issues.POST(
			"",
			handler.CreateIssue,
		)

		issues.GET(
			"",
			handler.ListIssues,
		)

		issues.PUT(
			"/:issueCode",
			handler.UpdateIssue,
		)

		issues.POST(
			"/:issueCode/resolveIssue",
			handler.ResolveIssue,
		)

		issues.GET(
			"/:issueCode/transactions",
			handler.ListIssueResolutionTransactions,
		)
	}
}
