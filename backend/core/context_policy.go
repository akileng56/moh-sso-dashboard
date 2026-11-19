package core

import (
	"context"
	"errors"
)

const (
	ContextNodeOnly           = "NODE_ONLY"
	ContextNodeAndDescendants = "NODE_AND_DESCENDANTS"
)

var ErrContextForbidden = errors.New("health context access denied")
var ErrContextResolverUnavailable = errors.New("context resolver is not configured")

// IDs and explanation details belong to the host. No UUID, SQL, geographic
// source or framework dependency is required by the portable policy.
type ContextAuthorization[ID comparable] struct {
	Subject            string
	Permissions        []string
	RequiredPermission string
	Systems            []string
	RequiredSystem     string
	ContextID          ID
}
type ContextDecision[E any] struct {
	Allowed bool
	Reason  string
	Details E
}
type ContextAccessResolver[ID comparable, E any] interface {
	ExplainContextAccess(context.Context, string, ID) (ContextDecision[E], error)
}

// AuthorizeContext checks exact permission/system matches in the existing order,
// then resolves the selected scope. Empty requirements are deliberately optional.
// The host authenticates subjects and resolves aliases before this operation.
func AuthorizeContext[ID comparable, E any](ctx context.Context, resolver ContextAccessResolver[ID, E], input ContextAuthorization[ID]) (ContextDecision[E], error) {
	if input.RequiredPermission != "" && !HasPermission(input.Permissions, input.RequiredPermission) {
		return ContextDecision[E]{Reason: "required functional permission is missing"}, ErrContextForbidden
	}
	// Unlike general launcher lookup, contextual system matching is exact.
	if input.RequiredSystem != "" && !HasPermission(input.Systems, input.RequiredSystem) {
		return ContextDecision[E]{Reason: "required system access is missing"}, ErrContextForbidden
	}
	if resolver == nil {
		return ContextDecision[E]{}, ErrContextResolverUnavailable
	}
	decision, err := resolver.ExplainContextAccess(ctx, input.Subject, input.ContextID)
	if err != nil {
		return ContextDecision[E]{}, err
	}
	if !decision.Allowed {
		return decision, ErrContextForbidden
	}
	return decision, nil
}

type ContextAssignment[ID comparable] struct {
	ContextID      ID
	ScopeMode      string
	AssignmentType string
}
type ContextHierarchy[ID comparable] interface {
	IsDescendant(context.Context, ID, ID) (bool, error)
}
type ContextScopeDecision struct {
	Allowed bool
	// AssignmentIndex refers to the original ordered effective-assignment list.
	// It is -1 on successful evaluations where no assignment matched.
	// On error the entire decision is zero; callers must check the error.
	AssignmentIndex int
	Reason          string
}

// ResolveContextScope consumes already-effective assignments. Expiry, group
// membership, defaults and enabled-assignment filtering remain provider duties.
// Same-node access does not query ancestry; inherited access requires the
// explicit descendant scope. Provider errors discard partial decisions.
func ResolveContextScope[ID comparable](ctx context.Context, hierarchy ContextHierarchy[ID], requested ID, enabled bool, assignments []ContextAssignment[ID]) (ContextScopeDecision, error) {
	denied := ContextScopeDecision{AssignmentIndex: -1}
	if !enabled {
		denied.Reason = "requested context is inactive"
		return denied, nil
	}
	for index, assignment := range assignments {
		allowed := assignment.ContextID == requested
		if !allowed && assignment.ScopeMode == ContextNodeAndDescendants {
			if hierarchy == nil {
				return ContextScopeDecision{}, ErrContextResolverUnavailable
			}
			var err error
			allowed, err = hierarchy.IsDescendant(ctx, assignment.ContextID, requested)
			if err != nil {
				return ContextScopeDecision{}, err
			}
		}
		if allowed {
			reason := "access granted by health context assignment"
			if assignment.AssignmentType == "DEFAULT" {
				reason = "access granted by the default National health context"
			}
			return ContextScopeDecision{Allowed: true, AssignmentIndex: index, Reason: reason}, nil
		}
	}
	denied.Reason = "no direct or inherited assignment covers this context"
	return denied, nil
}
