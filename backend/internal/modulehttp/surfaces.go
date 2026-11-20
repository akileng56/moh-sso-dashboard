package modulehttp

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
)

// DescribedModule owns both route inventory and its security declaration.
type DescribedModule interface {
	HTTPModule
	Patterns() []string
	OpenAPI() []byte
}

// RegisterSurfaces mounts only explicitly anonymous operations on the public
// group. All other declared operations retain platform authentication middleware.
// No application routes or permission decisions are duplicated in the host.
func RegisterSurfaces(public, protected *gin.RouterGroup, module DescribedModule) {
	var document struct {
		Paths map[string]map[string]struct{ Security json.RawMessage }
	}
	if err := json.Unmarshal(module.OpenAPI(), &document); err != nil {
		panic("invalid module security declaration")
	}
	for _, pattern := range module.Patterns() {
		method, path, ok := strings.Cut(pattern, " ")
		if !ok {
			panic("invalid module HTTP pattern")
		}
		operation, exists := document.Paths[path][strings.ToLower(method)]
		if !exists {
			panic("missing module security declaration")
		}
		group := protected
		// Missing/null/nonempty security always stays protected.
		var security []any
		if len(operation.Security) > 0 && string(operation.Security) != "null" {
			if err := json.Unmarshal(operation.Security, &security); err != nil {
				panic("invalid module security declaration")
			}
			if len(security) == 0 {
				group = public
			}
		}
		if group != nil {
			NewRegistrar(group).Handle(pattern, module)
		}
	}
}
