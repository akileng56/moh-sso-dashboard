package admin_units

import (
	"context"
	"errors"
	"reflect"
	"testing"

	units "github.com/mohuganda/moh-sso-dashboard/backend/modules/admin-units"
)

type adapterRepository struct {
	Repository
	ctx  context.Context
	rows []OrgUnit
	err  error
}

func (r *adapterRepository) ListHierarchy(ctx context.Context) ([]OrgUnit, error) {
	r.ctx = ctx
	return r.rows, r.err
}

func TestHostAdaptersUseSelectedPackageTypesAndBehavior(t *testing.T) {
	if reflect.TypeOf(OrgUnit{}).PkgPath() != units.ModulePath || ErrScopeForbidden != units.ErrScopeForbidden {
		t.Fatal("host retained a separate model or forbidden sentinel")
	}
	type requestKey struct{}
	ctx := context.WithValue(context.Background(), requestKey{}, "host-request")
	repo := &adapterRepository{rows: []OrgUnit{{DimOrgHierarchyKey: 1, Level: "1"}}}
	tree, err := NewService(repo).GetHierarchy(ctx)
	if err != nil || repo.ctx != ctx || len(tree) != 1 || tree[0].Name == nil || *tree[0].Name != "National" {
		t.Fatalf("service delegation changed: %+v %v", tree, err)
	}
	if !reflect.DeepEqual(tree, BuildHierarchyTree(repo.rows)) {
		t.Fatal("HTTP hierarchy delegation changed")
	}
	reader := &visibilityFixture{aliases: []VisibilityAlias{{Namespace: "dwh-org-unit", ExternalID: " x "}}}
	ids, scoped, err := ResolveVisibility(ctx, reader, "selected-id", true)
	if err != nil || !scoped || len(ids) != 1 || reader.ctx != ctx || reader.id != "selected-id" || !reader.descendants {
		t.Fatal("visibility adapter lost selected scope or request context")
	}
	cause := errors.New("warehouse unavailable")
	repo.err = cause
	if _, err := NewService(repo).GetHierarchy(ctx); !errors.Is(err, cause) {
		t.Fatal("published service error cause lost")
	}
}
