// Package core provides framework-independent platform contracts and the
// compiled-module lifecycle. HTTP, identity/database providers and business
// features are supplied by host adapters.
package core

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"time"
)

type ModuleInfo struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Enabled bool   `json:"enabled"`
}

type Stop func(context.Context) error

// ModuleDefinition describes an optional compiled module's lifecycle.
// Required platform capabilities are composed by the host, outside this list.
// Initialize must finish readiness before returning; partial acquisitions must
// return cleanup even on error. Cleanup cooperates with its context deadline.
type ModuleDefinition struct {
	ModuleInfo
	Requires   []string
	Initialize func(context.Context) (Stop, error)
}

type Runtime struct {
	definitions []ModuleDefinition
	mu          sync.Mutex
	initialized bool
}

var validID = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// NewRuntime validates explicit composition order before any module has side effects.
// Required modules must precede their consumers and be enabled when used.
func NewRuntime(definitions ...ModuleDefinition) (*Runtime, error) {
	seen := make(map[string]bool, len(definitions))
	for _, definition := range definitions {
		if !validID.MatchString(definition.ID) || definition.Version == "" {
			return nil, fmt.Errorf("module ID and version required: %q", definition.ID)
		}
		if _, exists := seen[definition.ID]; exists {
			return nil, fmt.Errorf("duplicate module %q", definition.ID)
		}
		for _, required := range definition.Requires {
			enabled, exists := seen[required]
			if !exists || (definition.Enabled && !enabled) {
				return nil, fmt.Errorf("module %q requires preceding enabled module %q", definition.ID, required)
			}
		}
		seen[definition.ID] = definition.Enabled
	}
	return &Runtime{definitions: append([]ModuleDefinition(nil), definitions...)}, nil
}

func (r *Runtime) Inventory() []ModuleInfo {
	items := make([]ModuleInfo, 0, len(r.definitions))
	for _, definition := range r.definitions {
		items = append(items, definition.ModuleInfo)
	}
	return items
}

// Initialize runs once. Failure cancels module contexts and rolls back acquired
// resources in reverse order. A failed registry must be discarded.
func (r *Runtime) Initialize(ctx context.Context) (Stop, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.initialized {
		return nil, errors.New("module registry already initialized")
	}
	r.initialized = true
	moduleCtx, cancel := context.WithCancel(ctx)
	var cleanups []Stop
	var once sync.Once
	var stopErr error
	stop := func(ctx context.Context) error {
		once.Do(func() {
			cancel()
			for i := len(cleanups) - 1; i >= 0; i-- {
				stopErr = errors.Join(stopErr, cleanups[i](ctx))
			}
		})
		return stopErr
	}
	for _, definition := range r.definitions {
		if !definition.Enabled || definition.Initialize == nil {
			continue
		}
		cleanup, err := definition.Initialize(moduleCtx)
		if cleanup != nil {
			cleanups = append(cleanups, cleanup)
		}
		if err != nil {
			rollbackCtx, rollbackCancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer rollbackCancel()
			return nil, errors.Join(fmt.Errorf("initialize module %q: %w", definition.ID, err), stop(rollbackCtx))
		}
	}
	return stop, nil
}
