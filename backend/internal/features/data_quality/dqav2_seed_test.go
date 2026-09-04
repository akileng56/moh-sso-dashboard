package data_quality

import (
	"strings"
	"testing"

	"github.com/moh-sso-dashboard/internal/features/data_quality/dqa"
)

// TestBuiltinRulesAllCompile proves every seeded rule (all ~52 of them, spread
// across row_expression, metric and raw_sql types) compiles to SQL without
// error. It cannot prove the SQL is semantically correct against a live
// table without a DWH connection, but it does catch typos, unbalanced
// parens, disallowed functions, and structurally invalid raw_sql — the class
// of mistake most likely in ~40 hand-written rule expressions.
func TestBuiltinRulesAllCompile(t *testing.T) {
	const tableID = "vht_monthly"
	const physical = "report.echis_vht_monthly"
	tables := map[string]string{tableID: physical}

	rules := BuiltinVHTMonthlyRules(tableID, physical)
	if len(rules) == 0 {
		t.Fatal("expected a non-empty builtin rule set")
	}

	seen := map[string]bool{}
	for _, r := range rules {
		if seen[r.Code] {
			t.Errorf("duplicate rule code: %s", r.Code)
		}
		seen[r.Code] = true

		if r.TableID != tableID {
			t.Errorf("rule %s: table_id = %q, want %q", r.Code, r.TableID, tableID)
		}

		compiled, err := dqa.CompileRule(r, tables, nil)
		if err != nil {
			t.Errorf("rule %s failed to compile: %v", r.Code, err)
			continue
		}
		if strings.TrimSpace(compiled.MeasureSQL) == "" {
			t.Errorf("rule %s produced empty measure SQL", r.Code)
		}
		if r.RuleType == dqa.RuleTypeRawSQL && !strings.Contains(compiled.MeasureSQL, physical) {
			t.Errorf("rule %s (raw_sql) does not reference the physical table %q: %s",
				r.Code, physical, compiled.MeasureSQL)
		}
	}

	t.Logf("compiled %d builtin rules successfully", len(rules))
}

// TestBuiltinRulesCoverAllCategories is a coarse sanity check that every
// category the original Python app documents is represented.
func TestBuiltinRulesCoverAllCategories(t *testing.T) {
	rules := BuiltinVHTMonthlyRules("t", "report.t")
	want := []string{
		"malaria", "diarrhoea", "pneumonia", "danger_signs", "pregnancy",
		"family_planning", "activity", "commodities", "outliers",
		"timeliness", "completeness", "validity", "consistency",
	}
	got := map[string]bool{}
	for _, r := range rules {
		got[r.Category] = true
	}
	for _, cat := range want {
		if !got[cat] {
			t.Errorf("no seeded rule in category %q", cat)
		}
	}
}
