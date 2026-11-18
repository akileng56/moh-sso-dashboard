package reportscheduler

import (
	"context"
	"testing"
)

func TestHealthHierarchyWithoutDatabaseFailsClosed(t *testing.T) {
	hierarchy := NewPostgresHealthHierarchy(nil)
	cases := []struct {
		name                     string
		requested, authenticated HealthContext
		want                     HealthContext
		wantError                bool
	}{
		{name: "national", want: HealthContext{Level: "national"}},
		{name: "fixed facility", authenticated: HealthContext{District: "Gulu", Facility: "facility-1"}, want: HealthContext{Level: "facility", District: "Gulu", Facility: "facility-1"}},
		{name: "fixed district", authenticated: HealthContext{District: "Gulu"}, want: HealthContext{Level: "district", District: "Gulu"}},
		{name: "different district", requested: HealthContext{District: "Kampala"}, authenticated: HealthContext{District: "Gulu"}, wantError: true},
		{name: "different facility", requested: HealthContext{Facility: "facility-2"}, authenticated: HealthContext{Facility: "facility-1"}, wantError: true},
		{name: "unverified district", requested: HealthContext{District: "Gulu"}, wantError: true},
		{name: "unverified facility", requested: HealthContext{Facility: "facility-1"}, wantError: true},
		{name: "unverified facility within district", requested: HealthContext{Facility: "facility-1"}, authenticated: HealthContext{District: "Gulu"}, wantError: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := hierarchy.Constrain(context.Background(), tc.requested, tc.authenticated)
			if (err != nil) != tc.wantError {
				t.Fatalf("scope=%+v error=%v", got, err)
			}
			if err == nil && got != tc.want {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

type rejectingHierarchy struct{}

func (rejectingHierarchy) Constrain(context.Context, HealthContext, HealthContext) (HealthContext, error) {
	return HealthContext{}, context.Canceled
}

func TestServiceUsesInjectedHealthHierarchy(t *testing.T) {
	s := NewService(nil, nil, nil, nil, nil, rejectingHierarchy{})
	if _, err := s.constrainHealthContext(context.Background(), HealthContext{}, HealthContext{}); err != context.Canceled {
		t.Fatalf("hierarchy error was not propagated: %v", err)
	}
}
