package admin_units

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/config"
	healthcontext "github.com/moh-sso-dashboard/internal/features/health_context"
)

func NewHandler(_ *config.Config, db *sql.DB, contexts *healthcontext.Service) *Handler {
	return NewHandlerWithPorts(NewService(NewRepository(db)), NewHealthContextVisibilityReader(contexts))
}

type healthContextVisibilityReader struct{ service *healthcontext.Service }

func NewHealthContextVisibilityReader(service *healthcontext.Service) VisibilityReader {
	if service == nil {
		return nil
	}
	return &healthContextVisibilityReader{service: service}
}
func visibilityHostError(err error) error {
	if errors.Is(err, healthcontext.ErrContextForbidden) {
		return errors.Join(ErrScopeForbidden, err)
	}
	return err
}
func (r *healthContextVisibilityReader) GetNode(ctx context.Context, id string) (VisibilityNode, error) {
	parsed, err := uuid.Parse(id)
	if err != nil || parsed == uuid.Nil {
		return VisibilityNode{}, ErrScopeForbidden
	}
	node, err := r.service.GetNode(ctx, parsed)
	return VisibilityNode{Code: node.Code}, visibilityHostError(err)
}
func (r *healthContextVisibilityReader) ListAliasesInScope(ctx context.Context, id string, descendants bool) ([]VisibilityAlias, error) {
	parsed, err := uuid.Parse(id)
	if err != nil || parsed == uuid.Nil {
		return nil, ErrScopeForbidden
	}
	aliases, err := r.service.ListAliasesInScope(ctx, parsed, descendants)
	if aliases == nil {
		return nil, visibilityHostError(err)
	}
	out := make([]VisibilityAlias, len(aliases))
	for i, alias := range aliases {
		out[i] = VisibilityAlias{Namespace: alias.Namespace, ExternalID: alias.ExternalID}
	}
	return out, visibilityHostError(err)
}

func (h *Handler) allowedOrgUnitIDs(c *gin.Context) (map[string]struct{}, bool, error) {
	raw, exists := c.Get("health_context_id")
	if !exists {
		return nil, false, nil
	}
	id, ok := raw.(uuid.UUID)
	if !ok || id == uuid.Nil {
		return nil, true, ErrScopeForbidden
	}
	mode, _ := c.Get("health_context_scope_mode")
	scopeMode, _ := mode.(healthcontext.ScopeMode)
	return ResolveVisibility(c.Request.Context(), h.visibility, id.String(), scopeMode == healthcontext.ScopeNodeAndDescendants)
}
