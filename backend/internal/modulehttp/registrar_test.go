package modulehttp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/moh-sso-dashboard/internal/authz"
	"github.com/moh-sso-dashboard/internal/observability"
)

func TestRegistrarPreservesVerifiedAccessTraceAndMountPath(t *testing.T) {
	for _, identity := range []string{"anonymous", "empty-subject", "denied", "allowed"} {
		t.Run(identity, func(t *testing.T) {
			outer := gin.New()
			group := outer.Group("/api/v1", func(c *gin.Context) {
				c.Request = c.Request.WithContext(observability.WithRequestID(c.Request.Context(), "trusted-trace"))
				if identity == "anonymous" {
					return
				}
				actor := authz.Context{UserID: "verified-subject"}
				if identity == "empty-subject" {
					actor.UserID = ""
				}
				if identity == "allowed" {
					actor.Permissions = []authz.Permission{"records:read"}
				}
				c.Set(authz.ContextKey, actor)
			})
			inner := http.NewServeMux()
			inner.HandleFunc("GET /records/{id}", func(w http.ResponseWriter, r *http.Request) {
				if err := AuthorizeAny(r, "records:read", "records:write"); err != nil {
					w.WriteHeader(err.(interface{ StatusCode() int }).StatusCode())
					return
				}
				fmt.Fprintf(w, "%s|%s", r.PathValue("id"), RequestID(r))
			})
			NewRegistrar(group).Handle("GET /records/{id}", inner)
			request := httptest.NewRequest("GET", "/api/v1/records/item-1", nil)
			request.Header.Set("Authorization", "Bearer unverified")
			request.Header.Set("X-User-ID", "spoofed")
			response := httptest.NewRecorder()
			outer.ServeHTTP(response, request)
			want := map[string]int{"anonymous": 401, "empty-subject": 401, "denied": 403, "allowed": 200}[identity]
			if response.Code != want {
				t.Fatalf("%s: %d", identity, response.Code)
			}
			if identity == "allowed" && response.Body.String() != "item-1|trusted-trace" {
				t.Fatal("path/context lost", response.Body.String())
			}
			head := httptest.NewRecorder()
			outer.ServeHTTP(head, httptest.NewRequest("HEAD", "/api/v1/records/item-1", nil))
			if head.Code != 404 {
				t.Fatal("registrar expanded the declared route surface", head.Code)
			}
		})
	}
}
func TestRegistrarRejectsUnsupportedPattern(t *testing.T) {
	for _, pattern := range []string{"", "/records", "GET /records/{bad-name}"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("invalid pattern accepted: %q", pattern)
				}
			}()
			NewRegistrar(gin.New().Group("/api/v1")).Handle(pattern, http.NotFoundHandler())
		}()
	}
}

type multipleRoutes struct{ *http.ServeMux }

func (m multipleRoutes) Patterns() []string { return []string{"GET /one", "POST /two"} }
func TestMultiplePatternsPreserveMethodSurfaceAndSystemBoundary(t *testing.T) {
	for _, system := range []bool{false, true} {
		engine := gin.New()
		group := engine.Group("/api/v1", func(c *gin.Context) {
			actor := authz.Context{UserID: "verified"}
			if system {
				actor.AccessibleSystems = []string{"system"}
			}
			c.Set(authz.ContextKey, actor)
		})
		inner := http.NewServeMux()
		handler := func(w http.ResponseWriter, r *http.Request) {
			if err := RequireSystem(r, "system"); err != nil {
				w.WriteHeader(err.(interface{ StatusCode() int }).StatusCode())
				return
			}
			w.WriteHeader(204)
		}
		inner.HandleFunc("GET /one", handler)
		inner.HandleFunc("POST /two", handler)
		Register(group, multipleRoutes{inner})
		for _, route := range []struct{ method, path string }{{"GET", "/one"}, {"POST", "/two"}, {"HEAD", "/one"}, {"GET", "/two"}} {
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, httptest.NewRequest(route.method, "/api/v1"+route.path, nil))
			want := 204
			if !system {
				want = 403
			}
			if route.method == "HEAD" || route.method == "GET" && route.path == "/two" {
				want = 404
			}
			if w.Code != want {
				t.Fatalf("%s %s: got %d want %d", route.method, route.path, w.Code, want)
			}
		}
	}
	if _, valid := Actor(httptest.NewRequest("GET", "/", nil)); valid {
		t.Fatal("unverified request has actor")
	}
	if err := RequireSystem(httptest.NewRequest("GET", "/", nil), "system"); err == nil {
		t.Fatal("unverified system granted")
	}
}
