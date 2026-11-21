package health_context

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/modulehttp"
)

type selectionFailure struct {
	status        int
	code, message string
}

func (e selectionFailure) Error() string         { return e.message }
func (e selectionFailure) StatusCode() int       { return e.status }
func (e selectionFailure) ErrorCode() string     { return e.code }
func (e selectionFailure) PublicMessage() string { return e.message }
func selectionError(status int, code, message string) error {
	return selectionFailure{status, code, message}
}

// resolveSelection is shared by Gin middleware and standard HTTP consumers.
func (s *Service) resolveSelection(ctx context.Context, userID, raw string) (uuid.UUID, AccessExplanation, error) {
	if s == nil || s.repository == nil {
		return uuid.Nil, AccessExplanation{}, selectionError(503, "HEALTH_CONTEXT_UNAVAILABLE", "health context access could not be verified")
	}
	value := strings.TrimSpace(raw)
	if value == "" {
		contexts, err := s.ListEffectiveContexts(ctx, userID)
		if err != nil {
			return uuid.Nil, AccessExplanation{}, selectionError(503, "HEALTH_CONTEXT_UNAVAILABLE", "health context access could not be verified")
		}
		if len(contexts) > 0 {
			return uuid.Nil, AccessExplanation{}, selectionError(400, "HEALTH_CONTEXT_REQUIRED", "select a health context for this request")
		}
		return uuid.Nil, AccessExplanation{}, nil
	}
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, AccessExplanation{}, selectionError(400, "INVALID_HEALTH_CONTEXT", "invalid health context")
	}
	explanation, err := s.ExplainAccess(ctx, userID, id)
	if err != nil || !explanation.Allowed {
		return uuid.Nil, explanation, selectionError(403, "HEALTH_CONTEXT_FORBIDDEN", "health context access denied")
	}
	return id, explanation, nil
}

// SelectContextHTTP applies the shared policy to a verified platform request.
func (s *Service) SelectContextHTTP(r *http.Request) (string, bool, error) {
	actor, valid := modulehttp.Actor(r)
	if !valid {
		return "", false, selectionError(401, "UNAUTHORIZED", "unauthorized")
	}
	id, explanation, err := s.resolveSelection(r.Context(), actor.UserID, r.Header.Get("X-Health-Context-ID"))
	if err != nil {
		return "", false, err
	}
	if id == uuid.Nil {
		return "", false, nil
	}
	return id.String(), explanation.ScopeMode == ScopeNodeAndDescendants, nil
}

// NodeCode and AliasIDs expose generic context capabilities, with no module DTOs.
func (s *Service) NodeCode(ctx context.Context, raw string) (string, error) {
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil || s == nil {
		return "", selectionError(403, "HEALTH_CONTEXT_FORBIDDEN", "health context access denied")
	}
	node, err := s.GetNode(ctx, id)
	return node.Code, contextCapabilityError(err)
}
func (s *Service) AliasIDs(ctx context.Context, raw string, descendants bool) (map[string][]string, error) {
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil || s == nil {
		return nil, selectionError(403, "HEALTH_CONTEXT_FORBIDDEN", "health context access denied")
	}
	aliases, err := s.ListAliasesInScope(ctx, id, descendants)
	if err != nil {
		return nil, contextCapabilityError(err)
	}
	result := map[string][]string{}
	for _, alias := range aliases {
		result[alias.Namespace] = append(result[alias.Namespace], alias.ExternalID)
	}
	return result, nil
}
func contextCapabilityError(err error) error {
	if errors.Is(err, ErrContextForbidden) {
		return selectionError(403, "HEALTH_CONTEXT_FORBIDDEN", "health context access denied")
	}
	return err
}
