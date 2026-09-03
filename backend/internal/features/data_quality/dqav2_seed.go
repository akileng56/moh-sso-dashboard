package data_quality

// BuiltinVHTMonthlyRules is a Go port of the original eCHIS DQA app's 37
// pandas-based checks (dqa/checks/*.py in the Python project), re-expressed
// as declarative v2 rules so they run through the same pushdown-SQL engine
// as user-authored rules.
//
// Column names match the Python app's documented defaults (dqa/columns.py,
// dqa/config.py ColumnConfig) exactly — these are the physical columns the
// original eCHIS VHT monthly extract uses. If your deployment's table uses
// different names, edit the generated rules after seeding (same manual step
// the original app requires via its own config.yaml `columns:` overrides).
//
// Most checks are simple per-row comparisons and map onto row_expression
// rules. The Python checks are per-sex (run once for female columns, once
// for male); each is collapsed here into ONE rule per code that flags a row
// if EITHER sex violates (OR), since a v2 rule produces one flag per
// violating row rather than the Python engine's one flag per (row, sex).
//
// A handful of checks are inherently sequential/statistical (consecutive
// months, month-over-month spikes, per-district percentiles, duplicate
// detection) and cannot be expressed as a single-row boolean predicate —
// SQL forbids window functions in a WHERE clause. These use the raw_sql
// escape hatch with a windowed CTE, which is exactly what raw_sql exists
// for. TIM-02 ("fewer than 1 period on record") is a no-op at the default
// threshold (a row's own group always has >=1 period) and is not seeded.

import "github.com/moh-sso-dashboard/internal/features/data_quality/dqa"

func neg(bad string) string { return "NOT (" + bad + ")" }

func rowRule(tableID, code, category, name, bad string) dqa.Rule {
	return dqa.Rule{
		Code: code, TableID: tableID, RuleType: dqa.RuleTypeRowExpression,
		Category: category, Name: name, Enabled: true, Source: "db",
		Zones:      dqa.Zone{Fail: "when > 0"},
		Definition: map[string]any{"expr": neg(bad)},
	}
}

func rawRule(tableID, code, category, name, failQuery string, sampleLimit int) dqa.Rule {
	def := map[string]any{"fail_query": failQuery}
	if sampleLimit > 0 {
		def["sample_limit"] = sampleLimit
	}
	return dqa.Rule{
		Code: code, TableID: tableID, RuleType: dqa.RuleTypeRawSQL,
		Category: category, Name: name, Enabled: true, Source: "db",
		Zones:      dqa.Zone{Fail: "when > 0"},
		Definition: def,
	}
}

// totalEventsExpr sums the same 16 "activity" columns the Python app uses
// for ACT-01..05, COMP-02 and the OUT-02..05 outlier checks.
const totalEventsExpr = `(` +
	`assess_u5_female + assess_u5_male + ` +
	`num_diarrhea_u5_female + num_diarrhea_u5_male + ` +
	`num_fever_u5_female + num_fever_u5_male + ` +
	`num_u5_with_pneumonia_female + num_u5_with_pneumonia_male + ` +
	`num_sick_reffered_female + num_sick_reffered_male + ` +
	`num_newborns_visited_twice_in_first_week_female + num_newborns_visited_twice_in_first_week_male + ` +
	`using_fp_method_female + using_fp_method_male + ` +
	`atleast_4_anc_visits + atleast_8_anc_visits` +
	`)`

// BuiltinVHTMonthlyRules returns every seeded rule for tableID, with raw_sql
// fail_query text pointing at physicalTable (raw_sql is not table-map
// substituted, so the physical name is embedded directly, same requirement
// the Python app has via its DB_TABLE env var).
func BuiltinVHTMonthlyRules(tableID, physicalTable string) []dqa.Rule {
	var rules []dqa.Rule

	// ── MAL-01..07: malaria cascade (per sex, collapsed via OR) ──────────
	malSex := func(f, m string) string { return "(" + f + ") OR (" + m + ")" }
	rules = append(rules,
		rowRule(tableID, "MAL-01", "malaria", "MRDT tested exceeds fever cases", malSex(
			"num_u5_with_fever_recieved_mrdt_female > num_fever_u5_female AND num_fever_u5_female > 0",
			"num_u5_with_fever_recieved_mrdt_male > num_fever_u5_male AND num_fever_u5_male > 0")),
		rowRule(tableID, "MAL-02", "malaria", "Confirmed malaria exceeds MRDT tested", malSex(
			"num_malaria_u5_female > num_u5_with_fever_recieved_mrdt_female AND num_u5_with_fever_recieved_mrdt_female > 0",
			"num_malaria_u5_male > num_u5_with_fever_recieved_mrdt_male AND num_u5_with_fever_recieved_mrdt_male > 0")),
		rowRule(tableID, "MAL-03", "malaria", "ACT dispensed exceeds confirmed malaria", malSex(
			"num_u5_malaria_recieved_act_female > num_malaria_u5_female AND num_malaria_u5_female > 0",
			"num_u5_malaria_recieved_act_male > num_malaria_u5_male AND num_malaria_u5_male > 0")),
		rowRule(tableID, "MAL-04", "malaria", "Low MRDT testing rate (<80%)", malSex(
			"num_fever_u5_female > 0 AND (num_u5_with_fever_recieved_mrdt_female::numeric / NULLIF(num_fever_u5_female,0)) < 0.80",
			"num_fever_u5_male > 0 AND (num_u5_with_fever_recieved_mrdt_male::numeric / NULLIF(num_fever_u5_male,0)) < 0.80")),
		rowRule(tableID, "MAL-05", "malaria", "Low ACT treatment rate (<80%)", malSex(
			"num_malaria_u5_female > 0 AND (num_u5_malaria_recieved_act_female::numeric / NULLIF(num_malaria_u5_female,0)) < 0.80",
			"num_malaria_u5_male > 0 AND (num_u5_malaria_recieved_act_male::numeric / NULLIF(num_malaria_u5_male,0)) < 0.80")),
		rowRule(tableID, "MAL-06", "malaria", "Danger signs but zero rectal artesunate", malSex(
			"num_u5_has_fever_and_danger_signs_female > 0 AND num_with_fever_and_danger_sign_treated_with_rectal_female = 0",
			"num_u5_has_fever_and_danger_signs_male > 0 AND num_with_fever_and_danger_sign_treated_with_rectal_male = 0")),
		rowRule(tableID, "MAL-07", "malaria", "Rectal artesunate exceeds danger sign cases", malSex(
			"num_with_fever_and_danger_sign_treated_with_rectal_female > num_u5_has_fever_and_danger_signs_female AND num_u5_has_fever_and_danger_signs_female > 0",
			"num_with_fever_and_danger_sign_treated_with_rectal_male > num_u5_has_fever_and_danger_signs_male AND num_u5_has_fever_and_danger_signs_male > 0")),
	)

	// ── DIA-01..05: diarrhoea cascade ─────────────────────────────────────
	diaSex := malSex
	rules = append(rules,
		rowRule(tableID, "DIA-01", "diarrhoea", "ORS treated exceeds diarrhoea episodes", diaSex(
			"num_diarrhea_u5_female_treated_with_ors > num_diarrhea_u5_female AND num_diarrhea_u5_female > 0",
			"num_diarrhea_u5_male_treated_with_ors > num_diarrhea_u5_male AND num_diarrhea_u5_male > 0")),
		rowRule(tableID, "DIA-02", "diarrhoea", "24hr treatment exceeds diarrhoea episodes", diaSex(
			"num_diarrhea_treated_24hrs_female > num_diarrhea_u5_female AND num_diarrhea_u5_female > 0",
			"num_diarrhea_treated_24hrs_male > num_diarrhea_u5_male AND num_diarrhea_u5_male > 0")),
		rowRule(tableID, "DIA-03", "diarrhoea", "24hr treatment exceeds ORS treated", diaSex(
			"num_diarrhea_treated_24hrs_female > num_diarrhea_u5_female_treated_with_ors AND num_diarrhea_u5_female_treated_with_ors > 0",
			"num_diarrhea_treated_24hrs_male > num_diarrhea_u5_male_treated_with_ors AND num_diarrhea_u5_male_treated_with_ors > 0")),
		rowRule(tableID, "DIA-04", "diarrhoea", "Low ORS treatment rate (<80%)", diaSex(
			"num_diarrhea_u5_female > 0 AND (num_diarrhea_u5_female_treated_with_ors::numeric / NULLIF(num_diarrhea_u5_female,0)) < 0.80",
			"num_diarrhea_u5_male > 0 AND (num_diarrhea_u5_male_treated_with_ors::numeric / NULLIF(num_diarrhea_u5_male,0)) < 0.80")),
		rowRule(tableID, "DIA-05", "diarrhoea", "Diarrhoea cases but zero ORS", diaSex(
			"num_diarrhea_u5_female > 0 AND num_diarrhea_u5_female_treated_with_ors = 0",
			"num_diarrhea_u5_male > 0 AND num_diarrhea_u5_male_treated_with_ors = 0")),
	)

	// ── PNE-01..05: pneumonia cascade ─────────────────────────────────────
	pneSex := malSex
	rules = append(rules,
		rowRule(tableID, "PNE-01", "pneumonia", "Amoxicillin exceeds pneumonia cases", pneSex(
			"num_u5_with_pneumonia_and_recieved_amoxicilin_female > num_u5_with_pneumonia_female AND num_u5_with_pneumonia_female > 0",
			"num_u5_with_pneumonia_and_recieved_amoxicilin_male > num_u5_with_pneumonia_male AND num_u5_with_pneumonia_male > 0")),
		rowRule(tableID, "PNE-02", "pneumonia", "24hr treatment exceeds pneumonia cases", pneSex(
			"num_u5_with_pneumonia_24hr_treatment_female > num_u5_with_pneumonia_female AND num_u5_with_pneumonia_female > 0",
			"num_u5_with_pneumonia_24hr_treatment_male > num_u5_with_pneumonia_male AND num_u5_with_pneumonia_male > 0")),
		rowRule(tableID, "PNE-03", "pneumonia", "24hr treatment exceeds amoxicillin treated", pneSex(
			"num_u5_with_pneumonia_24hr_treatment_female > num_u5_with_pneumonia_and_recieved_amoxicilin_female AND num_u5_with_pneumonia_and_recieved_amoxicilin_female > 0",
			"num_u5_with_pneumonia_24hr_treatment_male > num_u5_with_pneumonia_and_recieved_amoxicilin_male AND num_u5_with_pneumonia_and_recieved_amoxicilin_male > 0")),
		rowRule(tableID, "PNE-04", "pneumonia", "Low amoxicillin treatment rate (<80%)", pneSex(
			"num_u5_with_pneumonia_female > 0 AND (num_u5_with_pneumonia_and_recieved_amoxicilin_female::numeric / NULLIF(num_u5_with_pneumonia_female,0)) < 0.80",
			"num_u5_with_pneumonia_male > 0 AND (num_u5_with_pneumonia_and_recieved_amoxicilin_male::numeric / NULLIF(num_u5_with_pneumonia_male,0)) < 0.80")),
		rowRule(tableID, "PNE-05", "pneumonia", "Pneumonia cases but zero amoxicillin", pneSex(
			"num_u5_with_pneumonia_female > 0 AND num_u5_with_pneumonia_and_recieved_amoxicilin_female = 0",
			"num_u5_with_pneumonia_male > 0 AND num_u5_with_pneumonia_and_recieved_amoxicilin_male = 0")),
	)

	// ── REF-01..04: danger signs & referral (sexes summed, not OR'd — the
	// Python check sums female+male into one combined value per row) ──────
	rules = append(rules,
		rowRule(tableID, "REF-01", "danger_signs", "Referred exceeds assessed",
			"(num_sick_reffered_female+num_sick_reffered_male) > (assess_u5_female+assess_u5_male) AND (assess_u5_female+assess_u5_male) > 0"),
		rowRule(tableID, "REF-02", "danger_signs", "Danger signs but zero referrals",
			"(num_u5_has_fever_and_danger_signs_female+num_u5_has_fever_and_danger_signs_male) > 0 AND (num_sick_reffered_female+num_sick_reffered_male) = 0"),
		rowRule(tableID, "REF-03", "danger_signs", "Recovered exceeds assessed",
			"(num_u5_recovered_after_vht_treatment_female+num_u5_recovered_after_vht_treatment_male) > (assess_u5_female+assess_u5_male) AND (assess_u5_female+assess_u5_male) > 0"),
		rowRule(tableID, "REF-04", "danger_signs", "High referral rate (>50% of assessed)",
			"(assess_u5_female+assess_u5_male) > 0 AND ((num_sick_reffered_female+num_sick_reffered_male)::numeric / NULLIF(assess_u5_female+assess_u5_male,0)) > 0.50"),
	)

	// ── ANC-01..03: pregnancy / ANC ────────────────────────────────────────
	rules = append(rules,
		rowRule(tableID, "ANC-01", "pregnancy", "8+ ANC visits exceeds 4+ ANC visits",
			"atleast_8_anc_visits > atleast_4_anc_visits AND atleast_4_anc_visits > 0"),
		rowRule(tableID, "ANC-02", "pregnancy", "ANC reported but VHT not expected to report",
			"(atleast_4_anc_visits > 0 OR atleast_8_anc_visits > 0) AND vhts_expected_to_report = 0"),
		rowRule(tableID, "ANC-03", "pregnancy", "Multiple maternal deaths in a single month",
			"died_during_pregnancy > 1"),
	)

	// ── FP-01: family planning (FP-02 is the district-percentile raw_sql) ─
	rules = append(rules,
		rowRule(tableID, "FP-01", "family_planning", "FP clients reported but VHT not implementing FP",
			"(using_fp_method_female + using_fp_method_male) > 0 AND vhts_implementing_fp = 0"),
	)

	// ── ACT-01,02,05: CHW activity (03/04 need window functions -> raw_sql)
	rules = append(rules,
		rowRule(tableID, "ACT-01", "activity", "Low monthly activity (<10 events)",
			"vhts_submitted_a_report = 1 AND "+totalEventsExpr+" > 0 AND "+totalEventsExpr+" < 10"),
		rowRule(tableID, "ACT-02", "activity", "Zero events but report was submitted",
			"vhts_submitted_a_report = 1 AND "+totalEventsExpr+" = 0"),
		rowRule(tableID, "ACT-05", "activity", "Expected to report but did not submit",
			"vhts_expected_to_report = 1 AND vhts_submitted_a_report = 0"),
	)

	// ── COM-01..05: commodity stock-out contradictions (COM-06 -> raw_sql) ─
	rules = append(rules,
		rowRule(tableID, "COM-01", "commodities", "ACT dispensed during ACT stock-out",
			"(num_u5_malaria_recieved_act_female+num_u5_malaria_recieved_act_male) > 0 AND act_stock_outs = 1"),
		rowRule(tableID, "COM-02", "commodities", "Amoxicillin dispensed during stock-out",
			"(num_u5_with_pneumonia_and_recieved_amoxicilin_female+num_u5_with_pneumonia_and_recieved_amoxicilin_male) > 0 AND amoxicillin_stock_outs = 1"),
		rowRule(tableID, "COM-03", "commodities", "ORS dispensed during ORS/zinc stock-out",
			"(num_diarrhea_u5_female_treated_with_ors+num_diarrhea_u5_male_treated_with_ors) > 0 AND ors_zinc_stock_outs = 1"),
		rowRule(tableID, "COM-04", "commodities", "MRDT used during MRDT stock-out",
			"(num_u5_with_fever_recieved_mrdt_female+num_u5_with_fever_recieved_mrdt_male) > 0 AND malaria_rdts_stock_outs = 1"),
		rowRule(tableID, "COM-05", "commodities", "Treatment recorded during full stock-out (all 4 commodities)",
			"((num_u5_malaria_recieved_act_female+num_u5_malaria_recieved_act_male)"+
				"+(num_u5_with_pneumonia_and_recieved_amoxicilin_female+num_u5_with_pneumonia_and_recieved_amoxicilin_male)"+
				"+(num_diarrhea_u5_female_treated_with_ors+num_diarrhea_u5_male_treated_with_ors)"+
				"+(num_u5_with_fever_recieved_mrdt_female+num_u5_with_fever_recieved_mrdt_male)) > 0"+
				" AND act_stock_outs=1 AND amoxicillin_stock_outs=1 AND ors_zinc_stock_outs=1 AND malaria_rdts_stock_outs=1"),
	)

	// ── COMP-01,02: completeness ────────────────────────────────────────
	rules = append(rules,
		rowRule(tableID, "COMP-01", "completeness", "Missing value in a key identity column",
			"vht_uuid IS NULL OR period_date IS NULL OR district IS NULL OR name IS NULL"),
		rowRule(tableID, "COMP-02", "completeness", "Zero activity across all indicator columns",
			totalEventsExpr+" = 0"),
	)

	// ── VAL-01: validity ─────────────────────────────────────────────────
	rules = append(rules,
		rowRule(tableID, "VAL-01", "validity", "Negative value in an activity column",
			"assess_u5_female < 0 OR assess_u5_male < 0 OR num_diarrhea_u5_female < 0 OR num_diarrhea_u5_male < 0"+
				" OR num_fever_u5_female < 0 OR num_fever_u5_male < 0 OR num_u5_with_pneumonia_female < 0 OR num_u5_with_pneumonia_male < 0"+
				" OR num_sick_reffered_female < 0 OR num_sick_reffered_male < 0"),
	)

	// ── CON-01..03: consistency / cascade integrity ───────────────────────
	rules = append(rules,
		rowRule(tableID, "CON-01", "consistency", "Negative assessed U5 counts",
			"assess_u5_female < 0 OR assess_u5_male < 0"),
		rowRule(tableID, "CON-02", "consistency", "Diarrhoea cascade broken (ORS/24hr exceed diarrhoea)",
			"(num_diarrhea_u5_female+num_diarrhea_u5_male) > 0 AND ("+
				"(num_diarrhea_u5_female_treated_with_ors+num_diarrhea_u5_male_treated_with_ors) > (num_diarrhea_u5_female+num_diarrhea_u5_male)"+
				" OR (num_diarrhea_treated_24hrs_female+num_diarrhea_treated_24hrs_male) > (num_diarrhea_u5_female+num_diarrhea_u5_male)"+
				" OR (num_diarrhea_treated_24hrs_female+num_diarrhea_treated_24hrs_male) > (num_diarrhea_u5_female_treated_with_ors+num_diarrhea_u5_male_treated_with_ors))"),
		rowRule(tableID, "CON-03", "consistency", "Fever/malaria cascade broken (MRDT/malaria/ACT exceed fever)",
			"(num_fever_u5_female+num_fever_u5_male) > 0 AND ("+
				"(num_u5_with_fever_recieved_mrdt_female+num_u5_with_fever_recieved_mrdt_male) > (num_fever_u5_female+num_fever_u5_male)"+
				" OR (num_malaria_u5_female+num_malaria_u5_male) > (num_u5_with_fever_recieved_mrdt_female+num_u5_with_fever_recieved_mrdt_male)"+
				" OR (num_u5_malaria_recieved_act_female+num_u5_malaria_recieved_act_male) > (num_malaria_u5_female+num_malaria_u5_male))"),
	)

	// ── ACT-03: inactive for 3+ consecutive months (gaps-and-islands) ─────
	rules = append(rules, rawRule(tableID, "ACT-03", "activity",
		"Inactive for 3+ consecutive months",
		`WITH act AS (
			SELECT vht_uuid, period_date, name, district, `+totalEventsExpr+` AS total_events
			FROM `+physicalTable+`
		), tagged AS (
			SELECT *, CASE WHEN total_events = 0 THEN 1 ELSE 0 END AS is_zero
			FROM act
		), grp AS (
			SELECT *,
			  ROW_NUMBER() OVER (PARTITION BY vht_uuid ORDER BY period_date)
			  - ROW_NUMBER() OVER (PARTITION BY vht_uuid, is_zero ORDER BY period_date) AS island
			FROM tagged
		), runs AS (
			SELECT *, COUNT(*) OVER (PARTITION BY vht_uuid, island) AS run_len
			FROM grp WHERE is_zero = 1
		)
		SELECT vht_uuid, period_date, name, district, run_len FROM runs WHERE run_len >= 3`, 200))

	// ── ACT-04: overreporting, top 5% by volume per district ──────────────
	rules = append(rules, rawRule(tableID, "ACT-04", "activity",
		"High volume: top 5% of district by monthly activity",
		`WITH act AS (
			SELECT vht_uuid, period_date, name, district, `+totalEventsExpr+` AS total_events
			FROM `+physicalTable+`
		), thresh AS (
			SELECT district, percentile_cont(0.95) WITHIN GROUP (ORDER BY total_events) AS p95
			FROM act WHERE total_events > 0 GROUP BY district
		)
		SELECT a.vht_uuid, a.period_date, a.name, a.district, a.total_events, t.p95
		FROM act a JOIN thresh t ON t.district = a.district
		WHERE a.total_events > 0 AND a.total_events > t.p95`, 200))

	// ── COM-06: chronic (3+ month) stock-out, one rule per commodity ──────
	chronic := func(code, label, col string) dqa.Rule {
		return rawRule(tableID, code, "commodities", "Chronic "+label+" stock-out (3+ consecutive months)",
			`WITH so AS (
				SELECT vht_uuid, period_date, name, district, `+col+` AS stocked_out
				FROM `+physicalTable+`
			), grp AS (
				SELECT *,
				  ROW_NUMBER() OVER (PARTITION BY vht_uuid ORDER BY period_date)
				  - ROW_NUMBER() OVER (PARTITION BY vht_uuid, stocked_out ORDER BY period_date) AS island
				FROM so
			), runs AS (
				SELECT *, COUNT(*) OVER (PARTITION BY vht_uuid, island) AS run_len
				FROM grp WHERE stocked_out = 1
			)
			SELECT vht_uuid, period_date, name, district, run_len FROM runs WHERE run_len >= 3`, 200)
	}
	rules = append(rules,
		chronic("COM-06-ACT", "ACT", "act_stock_outs"),
		chronic("COM-06-AMOX", "amoxicillin", "amoxicillin_stock_outs"),
		chronic("COM-06-ORS", "ORS/zinc", "ors_zinc_stock_outs"),
		chronic("COM-06-MRDT", "MRDT", "malaria_rdts_stock_outs"),
	)

	// ── OUT-02/03: month-over-month spike / drop (3x) ─────────────────────
	rules = append(rules, rawRule(tableID, "OUT-02", "outliers",
		"Month-over-month spike (>3x previous)",
		`WITH act AS (
			SELECT vht_uuid, period_date, name, district, `+totalEventsExpr+` AS total_events
			FROM `+physicalTable+`
		), prev AS (
			SELECT *, LAG(total_events) OVER (PARTITION BY vht_uuid ORDER BY period_date) AS prev_total
			FROM act
		)
		SELECT vht_uuid, period_date, name, district, total_events, prev_total FROM prev
		WHERE prev_total > 0 AND total_events > 3 * prev_total`, 200))

	rules = append(rules, rawRule(tableID, "OUT-03", "outliers",
		"Month-over-month drop (previous >3x current)",
		`WITH act AS (
			SELECT vht_uuid, period_date, name, district, `+totalEventsExpr+` AS total_events
			FROM `+physicalTable+`
		), prev AS (
			SELECT *, LAG(total_events) OVER (PARTITION BY vht_uuid ORDER BY period_date) AS prev_total
			FROM act
		)
		SELECT vht_uuid, period_date, name, district, total_events, prev_total FROM prev
		WHERE total_events > 0 AND prev_total > 3 * total_events`, 200))

	// ── OUT-04: identical non-zero total for 6+ consecutive months ────────
	rules = append(rules, rawRule(tableID, "OUT-04", "outliers",
		"Identical non-zero total for 6+ consecutive months",
		`WITH act AS (
			SELECT vht_uuid, period_date, name, district, `+totalEventsExpr+` AS total_events
			FROM `+physicalTable+`
		), prev AS (
			SELECT *, LAG(total_events) OVER (PARTITION BY vht_uuid ORDER BY period_date) AS prev_total
			FROM act
		), chained AS (
			SELECT *, SUM(CASE WHEN prev_total IS DISTINCT FROM total_events THEN 1 ELSE 0 END)
			          OVER (PARTITION BY vht_uuid ORDER BY period_date) AS chain_id
			FROM prev
		), runs AS (
			SELECT *, COUNT(*) OVER (PARTITION BY vht_uuid, chain_id) AS run_len
			FROM chained
		)
		SELECT vht_uuid, period_date, name, district, total_events, run_len FROM runs
		WHERE total_events > 0 AND run_len >= 6`, 200))

	// ── OUT-05: consecutive round numbers (modulo 5, min 10, 3+ months) ───
	rules = append(rules, rawRule(tableID, "OUT-05", "outliers",
		"Round numbers for 3+ consecutive months",
		`WITH act AS (
			SELECT vht_uuid, period_date, name, district, `+totalEventsExpr+` AS total_events
			FROM `+physicalTable+`
		), tagged AS (
			SELECT *, CASE WHEN total_events >= 10 AND total_events % 5 = 0 THEN 1 ELSE 0 END AS is_round
			FROM act
		), grp AS (
			SELECT *,
			  ROW_NUMBER() OVER (PARTITION BY vht_uuid ORDER BY period_date)
			  - ROW_NUMBER() OVER (PARTITION BY vht_uuid, is_round ORDER BY period_date) AS island
			FROM tagged
		), runs AS (
			SELECT *, COUNT(*) OVER (PARTITION BY vht_uuid, island) AS run_len
			FROM grp WHERE is_round = 1
		)
		SELECT vht_uuid, period_date, name, district, total_events, run_len FROM runs WHERE run_len >= 3`, 200))

	// ── TIM-01: reporting gap >1 month between consecutive periods ────────
	// (TIM-02, "fewer than 1 period on record", is inert at the default
	// min_periods_per_entity=1 threshold and is intentionally not seeded.)
	rules = append(rules, rawRule(tableID, "TIM-01", "timeliness",
		"Reporting gap of more than 1 month",
		`WITH ordered AS (
			SELECT vht_uuid, period_date, name, district,
			       LAG(period_date) OVER (PARTITION BY vht_uuid ORDER BY period_date) AS prev_period
			FROM `+physicalTable+`
		)
		SELECT vht_uuid, period_date, name, district, prev_period FROM ordered
		WHERE prev_period IS NOT NULL
		  AND (EXTRACT(YEAR FROM period_date::date)*12 + EXTRACT(MONTH FROM period_date::date))
			- (EXTRACT(YEAR FROM prev_period::date)*12 + EXTRACT(MONTH FROM prev_period::date)) > 1`, 200))

	// ── FP-02: FP count exceeds district 95th percentile ──────────────────
	rules = append(rules, rawRule(tableID, "FP-02", "family_planning",
		"FP count exceeds district 95th percentile",
		`WITH fp AS (
			SELECT vht_uuid, period_date, name, district,
			       (using_fp_method_female + using_fp_method_male) AS fp_total
			FROM `+physicalTable+`
		), thresh AS (
			SELECT district, percentile_cont(0.95) WITHIN GROUP (ORDER BY fp_total) AS p95
			FROM fp WHERE fp_total > 0 GROUP BY district
		)
		SELECT f.vht_uuid, f.period_date, f.name, f.district, f.fp_total, t.p95
		FROM fp f JOIN thresh t ON t.district = f.district
		WHERE f.fp_total > 0 AND f.fp_total > t.p95`, 200))

	// ── UNQ-01: duplicate (vht_uuid, period_date) ─────────────────────────
	rules = append(rules, rawRule(tableID, "UNQ-01", "uniqueness",
		"Duplicate (entity, period)",
		`SELECT a.vht_uuid, a.period_date, a.name, a.district
		 FROM `+physicalTable+` a
		 WHERE EXISTS (
			SELECT 1 FROM `+physicalTable+` b
			WHERE b.vht_uuid = a.vht_uuid AND b.period_date = a.period_date AND b.ctid <> a.ctid
		 )`, 200))

	return rules
}
