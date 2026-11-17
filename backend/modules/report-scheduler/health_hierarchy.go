package reportscheduler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// HealthHierarchy constrains requested scope against the authenticated scope.
// Implementations must reject unverifiable district/facility relationships.
type HealthHierarchy interface {
	Constrain(context.Context, HealthContext, HealthContext) (HealthContext, error)
}

type postgresHealthHierarchy struct{ dwh *sql.DB }

// NewPostgresHealthHierarchy reads the host-owned DWH hierarchy. A nil database
// still permits fixed authenticated scope; scope expansion fails closed.
func NewPostgresHealthHierarchy(db *sql.DB) HealthHierarchy {
	return &postgresHealthHierarchy{dwh: db}
}

func (s *postgresHealthHierarchy) Constrain(ctx context.Context, requested, user HealthContext) (HealthContext, error) {
	requested.Level = strings.ToLower(strings.TrimSpace(requested.Level))
	requested.District = strings.TrimSpace(requested.District)
	requested.Facility = strings.TrimSpace(requested.Facility)
	user.District = strings.TrimSpace(user.District)
	user.Facility = strings.TrimSpace(user.Facility)

	if user.Facility != "" {
		if requested.Facility != "" && !strings.EqualFold(requested.Facility, user.Facility) {
			return requested, errors.New("requested facility is outside the authenticated user's health context")
		}
		if user.District != "" && requested.District != "" && !strings.EqualFold(requested.District, user.District) {
			return requested, errors.New("requested district is outside the authenticated user's health context")
		}
		requested.Facility = user.Facility
		requested.District = user.District
		requested.Level = "facility"
		return requested, nil
	}
	if user.District != "" {
		if requested.District != "" && !strings.EqualFold(requested.District, user.District) {
			return requested, errors.New("requested district is outside the authenticated user's health context")
		}
		if requested.Facility != "" {
			if s.dwh == nil {
				return requested, errors.New("facility scope validation requires the DWH connection")
			}
			var exists bool
			err := s.dwh.QueryRowContext(ctx, `SELECT EXISTS (
				SELECT 1 FROM dwh.dim_org_hierarchy
				WHERE is_current=true AND "level"='6'
				  AND (facility_uid=$1 OR facility_name=$1 OR org_unit_id=$1 OR org_unit_name=$1)
				  AND (district_uid=$2 OR district=$2)
			)`, requested.Facility, user.District).Scan(&exists)
			if err != nil {
				return requested, fmt.Errorf("validate facility health context: %w", err)
			}
			if !exists {
				return requested, errors.New("requested facility is outside the authenticated user's district")
			}
			requested.District = user.District
			requested.Level = "facility"
			return requested, nil
		}
		requested.District = user.District
		requested.Level = "district"
		return requested, nil
	}
	if requested.Facility != "" {
		if s.dwh == nil {
			return requested, errors.New("facility scope validation requires the DWH connection")
		}
		var district sql.NullString
		err := s.dwh.QueryRowContext(ctx, `SELECT district
			FROM dwh.dim_org_hierarchy
			WHERE is_current=true AND "level"='6'
			  AND (facility_uid=$1 OR facility_name=$1 OR org_unit_id=$1 OR org_unit_name=$1)
			LIMIT 1`, requested.Facility).Scan(&district)
		if errors.Is(err, sql.ErrNoRows) {
			return requested, errors.New("requested facility was not found in the health hierarchy")
		}
		if err != nil {
			return requested, fmt.Errorf("validate facility health context: %w", err)
		}
		if requested.District != "" && district.Valid && !strings.EqualFold(requested.District, district.String) {
			return requested, errors.New("requested facility does not belong to the requested district")
		}
		if requested.District == "" && district.Valid {
			requested.District = district.String
		}
		requested.Level = "facility"
		return requested, nil
	}
	if requested.District != "" {
		if s.dwh == nil {
			return requested, errors.New("district scope validation requires the DWH connection")
		}
		var exists bool
		err := s.dwh.QueryRowContext(ctx, `SELECT EXISTS (
			SELECT 1 FROM dwh.dim_org_hierarchy
			WHERE is_current=true AND (district_uid=$1 OR district=$1)
		)`, requested.District).Scan(&exists)
		if err != nil {
			return requested, fmt.Errorf("validate district health context: %w", err)
		}
		if !exists {
			return requested, errors.New("requested district was not found in the health hierarchy")
		}
		requested.Level = "district"
		return requested, nil
	}
	requested.Level = "national"
	return requested, nil
}
