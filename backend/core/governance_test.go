package core

import (
	"errors"
	"testing"
)

func TestApplicationLifecycleMatrixAndReviewAuthority(t *testing.T) {
	edges := map[[2]string]bool{
		{"registered", "configured"}: true, {"registered", "retired"}: true,
		{"configured", "validated"}: true, {"configured", "retired"}: true,
		{"validated", "approved"}: true, {"validated", "configured"}: true, {"validated", "retired"}: true,
		{"approved", "active"}: true, {"approved", "configured"}: true, {"approved", "retired"}: true,
		{"active", "suspended"}: true, {"active", "retired"}: true,
		{"suspended", "configured"}: true, {"suspended", "retired"}: true,
	}
	statuses := append(ApplicationLifecycleStatuses(), "unknown", "ACTIVE", "")
	for _, from := range statuses {
		for _, to := range statuses {
			if got := CanTransitionApplication(from, to); got != edges[[2]string{from, to}] {
				t.Fatalf("transition %q -> %q: %v", from, to, got)
			}
		}
	}
	for _, target := range statuses {
		for _, subject := range []string{"", "reviewer"} {
			for _, configure := range []bool{false, true} {
				for _, approve := range []bool{false, true} {
					want := IsApplicationLifecycleStatus(target) && subject != "" && configure && ((target != "approved" && target != "active") || approve)
					if got := CanManageApplicationLifecycle(LifecycleActor{subject, configure, approve}, target); got != want {
						t.Fatalf("incorrect review authority for %q", target)
					}
				}
			}
		}
	}
	// A caller cannot mutate the authoritative state machine through the listing.
	copy := ApplicationLifecycleStatuses()
	copy[0] = "active"
	if ApplicationLifecycleStatuses()[0] != "registered" {
		t.Fatal("mutable status catalogue")
	}
}

func TestApplicationMetadataValidation(t *testing.T) {
	valid := Application{
		Access:    SystemAccess{DisplayName: "Pilot", LaunchURL: "https://example.org", SystemType: "external"},
		OwnerTeam: "team", OwnerName: "Owner", OwnerEmail: "owner@example.org", SupportURL: "/support",
		Environment: "test", ContractVersion: "1", IntegrationType: "external",
	}
	if err := ValidateApplicationMetadata(valid); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		change  func(*Application)
		message string
	}{
		{"required order", func(a *Application) { a.Access.DisplayName = " "; a.OwnerTeam = "" }, "displayName is required before validation"},
		{"email", func(a *Application) { a.OwnerEmail = "Owner <owner@example.org>" }, "ownerEmail must be an email address"},
		{"support", func(a *Application) { a.SupportURL = "//example.org" }, "supportUrl must be an absolute URL or root-relative path"},
		{"documentation", func(a *Application) { a.DocumentationURL = "relative" }, "documentationUrl must be an absolute URL or root-relative path"},
		{"environment", func(a *Application) { a.Environment = "TEST" }, "environment must be development, test, staging, or production"},
		{"version", func(a *Application) { a.ContractVersion = "2" }, "unsupported contractVersion"},
		{"integration", func(a *Application) { a.Access.SystemType = "platform" }, "integrationType must match systemType"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := valid
			test.change(&a)
			err := ValidateApplicationMetadata(a)
			if !errors.Is(err, ErrInvalidApplicationMetadata) || err.Error() != test.message {
				t.Fatalf("got %v", err)
			}
		})
	}
	valid.Access.SystemType = "platform"
	valid.IntegrationType = "hosted_module"
	valid.Access.LaunchURL = "/apps/pilot"
	if err := ValidateApplicationMetadata(valid); err != nil {
		t.Fatal(err)
	}
}
