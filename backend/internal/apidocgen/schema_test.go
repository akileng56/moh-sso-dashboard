package apidocgen

import (
	"go/token"
	"go/types"
	"reflect"
	"testing"

	scheduler "github.com/mohuganda/moh-sso-dashboard/backend/modules/report-scheduler"
)

func TestRequiredPointerBooleanRejectsNullInContract(t *testing.T) {
	s := &schemas{definitions: object{}}
	field := types.NewVar(token.NoPos, nil, "Enabled", types.NewPointer(types.Typ[types.Bool]))
	for _, required := range []bool{false, true} {
		tag := `json:"enabled"`
		if required {
			tag += ` binding:"required"`
		}
		schema := s.of(types.NewStruct([]*types.Var{field}, []string{tag}))
		property := schema["properties"].(object)["enabled"].(object)
		if property["type"] != "boolean" {
			t.Fatalf("unexpected property: %v", property)
		}
		if required && property["nullable"] != nil {
			t.Fatalf("required pointer must reject null: %v", property)
		}
		if !required && property["nullable"] != true {
			t.Fatalf("optional pointer should allow null: %v", property)
		}
	}
}

func TestSchedulerExtractionPreservesPublicSchemaNames(t *testing.T) {
	schemasFor := func(path, name string) object {
		pkg := types.NewPackage(path, name)
		field := types.NewVar(token.NoPos, pkg, "ID", types.Typ[types.String])
		record := types.NewNamed(types.NewTypeName(token.NoPos, pkg, "Schedule", nil), types.NewStruct([]*types.Var{field}, []string{`json:"id"`}), nil)
		schemas := &schemas{definitions: object{}}
		return object{"reference": schemas.of(record), "definitions": schemas.definitions}
	}
	legacy := schemasFor("github.com/moh-sso-dashboard/internal/features/report_scheduler", "report_scheduler")
	extracted := schemasFor(scheduler.ModulePath, "reportscheduler")
	if !reflect.DeepEqual(legacy, extracted) {
		t.Fatalf("scheduler schema changed after extraction: old=%v new=%v", legacy, extracted)
	}
}
