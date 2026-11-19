package core

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type hierarchyFunc func(context.Context, string, string) (bool, error)

func (f hierarchyFunc) IsDescendant(ctx context.Context, a, b string) (bool, error) {
	return f(ctx, a, b)
}

type contextResolverFunc func(context.Context, string, string) (ContextDecision[string], error)

func (f contextResolverFunc) ExplainContextAccess(ctx context.Context, s, id string) (ContextDecision[string], error) {
	return f(ctx, s, id)
}

func TestContextScopeIsolationAndAssignmentOrder(t *testing.T) {
	ctx := context.Background()
	calls := 0
	hierarchy := hierarchyFunc(func(got context.Context, a, b string) (bool, error) {
		calls++
		if got != ctx || a != "national" || b != "district" {
			t.Fatalf("ancestry forwarded incorrectly: %s -> %s", a, b)
		}
		return true, nil
	})
	assignments := []ContextAssignment[string]{
		{"unrelated", ContextNodeOnly, "DIRECT"},
		{"national", ContextNodeAndDescendants, "DEFAULT"},
		{"district", ContextNodeOnly, "GROUP"},
	}
	decision, err := ResolveContextScope(ctx, hierarchy, "district", true, assignments)
	if err != nil || !decision.Allowed || decision.AssignmentIndex != 1 || decision.Reason != "access granted by the default National health context" || calls != 1 {
		t.Fatalf("incorrect inherited/default decision: %+v, %v, calls=%d", decision, err, calls)
	}
	// Disabled contexts and direct matches never consult the hierarchy provider.
	for _, enabled := range []bool{false, true} {
		decision, err = ResolveContextScope[string](ctx, nil, "district", enabled, assignments[2:])
		if err != nil || decision.Allowed != enabled {
			t.Fatalf("direct/inactive: %+v %v", decision, err)
		}
	}
	for _, mode := range []string{ContextNodeOnly, "", "NODE_AND_DESCENDANT"} {
		decision, err = ResolveContextScope[string](ctx, nil, "district", true, []ContextAssignment[string]{{"national", mode, "DIRECT"}})
		if err != nil || decision.Allowed || decision.AssignmentIndex != -1 {
			t.Fatalf("scope widened: %+v %v", decision, err)
		}
	}
	cause := errors.New("hierarchy unavailable")
	failure := hierarchyFunc(func(context.Context, string, string) (bool, error) { return true, cause })
	decision, err = ResolveContextScope(ctx, failure, "district", true, assignments)
	if !errors.Is(err, cause) || !reflect.DeepEqual(decision, ContextScopeDecision{}) {
		t.Fatalf("partial scope leaked: %+v %v", decision, err)
	}
	_, err = ResolveContextScope[string](ctx, nil, "district", true, assignments)
	if !errors.Is(err, ErrContextResolverUnavailable) {
		t.Fatalf("missing ancestry provider: %v", err)
	}
}

func TestContextAuthorizationGatesAndProviderFailure(t *testing.T) {
	ctx := context.WithValue(context.Background(), struct{}{}, "trace")
	input := ContextAuthorization[string]{Subject: "subject", ContextID: "district", RequiredPermission: "read", RequiredSystem: "data-statistics"}
	calls := 0
	resolver := contextResolverFunc(func(got context.Context, subject, id string) (ContextDecision[string], error) {
		calls++
		if got != ctx || subject != "subject" || id != "district" {
			t.Fatal("lost identity, selection or context")
		}
		return ContextDecision[string]{true, "assigned", "host explanation"}, nil
	})
	denied, err := AuthorizeContext[string, string](ctx, resolver, input)
	if !errors.Is(err, ErrContextForbidden) || calls != 0 || denied.Reason != "required functional permission is missing" {
		t.Fatal(denied, err)
	}
	input.Permissions = []string{"read"}
	for _, system := range []string{"Data-statistics", " data-statistics", "data-statistics "} {
		input.Systems = []string{system}
		denied, err = AuthorizeContext[string, string](ctx, resolver, input)
		if !errors.Is(err, ErrContextForbidden) || calls != 0 || denied.Reason != "required system access is missing" {
			t.Fatal(denied, err)
		}
	}
	input.Systems = []string{"data-statistics"}
	allowed, err := AuthorizeContext[string, string](ctx, resolver, input)
	if err != nil || !allowed.Allowed || allowed.Details != "host explanation" || calls != 1 {
		t.Fatal(allowed, err)
	}
	failure := errors.New("provider failure")
	resolver = func(context.Context, string, string) (ContextDecision[string], error) {
		return ContextDecision[string]{true, "partial grant", "sensitive partial details"}, failure
	}
	denied, err = AuthorizeContext[string, string](ctx, resolver, input)
	if !errors.Is(err, failure) || denied != (ContextDecision[string]{}) {
		t.Fatal("partial grant leaked", denied, err)
	}
	resolver = func(context.Context, string, string) (ContextDecision[string], error) {
		return ContextDecision[string]{false, "not assigned", "host denial"}, nil
	}
	denied, err = AuthorizeContext[string, string](ctx, resolver, input)
	if !errors.Is(err, ErrContextForbidden) || denied.Details != "host denial" || denied.Allowed {
		t.Fatal(denied, err)
	}
	_, err = AuthorizeContext[string, string](ctx, nil, input)
	if !errors.Is(err, ErrContextResolverUnavailable) {
		t.Fatal(err)
	}
	// Empty requirements remain optional; they do not skip the scope resolver.
	input.RequiredPermission = ""
	input.RequiredSystem = ""
	denied, err = AuthorizeContext[string, string](ctx, resolver, input)
	if !errors.Is(err, ErrContextForbidden) || denied.Reason != "not assigned" {
		t.Fatal(denied, err)
	}
}
