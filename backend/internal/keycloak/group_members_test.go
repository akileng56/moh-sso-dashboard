package keycloak

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

// newGroupMembersServer serves a group of `total` members, returning at most
// min(max, serverCap) users per request (serverCap <= 0 means no cap).
func newGroupMembersServer(t *testing.T, total int, serverCap int) *httptest.Server {
	t.Helper()
	members := make([]KeycloakUser, total)
	for i := range members {
		members[i] = KeycloakUser{
			ID:        fmt.Sprintf("u%d", i),
			Username:  fmt.Sprintf("user%d", i),
			Email:     fmt.Sprintf("user%d@example.org", i),
			FirstName: fmt.Sprintf("First%d", i),
			LastName:  fmt.Sprintf("Last%d", i),
		}
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/realms/test/protocol/openid-connect/token":
			_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "test-token", "expires_in": 300})
		case "/admin/realms/test/groups/g1/members":
			first, _ := strconv.Atoi(r.URL.Query().Get("first"))
			max, err := strconv.Atoi(r.URL.Query().Get("max"))
			switch {
			case err != nil || max == 0:
				max = 100 // Keycloak default when max is absent
			case max < 0:
				max = len(members) // Keycloak: negative max = no limit
			}
			if serverCap > 0 && max > serverCap {
				max = serverCap
			}
			if first > len(members) {
				first = len(members)
			}
			end := first + max
			if end > len(members) {
				end = len(members)
			}
			_ = json.NewEncoder(w).Encode(members[first:end])
		default:
			t.Errorf("unexpected request %s", r.URL.String())
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestListGroupMembersReturnsMembersBeyondFirstPage(t *testing.T) {
	for _, tc := range []struct {
		name      string
		total     int
		serverCap int
	}{
		{"exact multiple of page size", 300, 0},
		{"partial last page", 257, 0},
		{"server caps page below requested size", 257, 50},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := newGroupMembersServer(t, tc.total, tc.serverCap)
			defer server.Close()
			client := NewAdminClient(server.URL, "test", "admin", "secret")

			members, err := client.ListGroupMembers(context.Background(), "g1")
			if err != nil {
				t.Fatal(err)
			}
			if len(members) != tc.total {
				t.Fatalf("got %d members, want %d", len(members), tc.total)
			}

			found, err := client.ListGroupMembersFiltered(context.Background(), "g1", "last250 first250")
			if err != nil {
				t.Fatal(err)
			}
			if len(found) != 1 || found[0].ID != "u250" {
				t.Fatalf("search for member #250 returned %+v", found)
			}
		})
	}
}
