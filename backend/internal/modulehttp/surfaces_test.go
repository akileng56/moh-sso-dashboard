package modulehttp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type surfacesFixture struct {
	schema   []byte
	patterns []string
}

func (m surfacesFixture) OpenAPI() []byte                                { return m.schema }
func (m surfacesFixture) Patterns() []string                             { return m.patterns }
func (surfacesFixture) ServeHTTP(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(204) }
func TestSurfacesRequireExplicitAnonymousDeclaration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := surfacesFixture{schema: []byte(`{"paths":{"/public":{"get":{"security":[]}},"/private":{"get":{}},"/null":{"get":{"security":null}}}}`),
		patterns: []string{"GET /public", "GET /private", "GET /null"}}
	engine := gin.New()
	public := engine.Group("/api/v1")
	private := engine.Group("/api/v1", func(c *gin.Context) { c.AbortWithStatus(401) })
	RegisterSurfaces(public, private, m)
	for _, path := range []string{"/public", "/private", "/null"} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1"+path, nil)
		req.Header.Set("X-User-ID", "spoofed")
		engine.ServeHTTP(w, req)
		want := 401
		if path == "/public" {
			want = 204
		}
		if w.Code != want {
			t.Fatalf("%s: %d", path, w.Code)
		}
	}
}
func TestSurfacesRejectInvalidOrMissingDeclarations(t *testing.T) {
	for _, schema := range []string{"{", `{"paths":{}}`, `{"paths":{"/private":{"get":{"security":"invalid"}}}}`} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid declaration accepted")
				}
			}()
			engine := gin.New()
			RegisterSurfaces(engine.Group("/public"), engine.Group("/protected"),
				surfacesFixture{schema: []byte(schema), patterns: []string{"GET /private"}})
		}()
	}
}
