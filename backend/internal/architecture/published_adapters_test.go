package architecture

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Published domains must have exactly one implementation owner. The host may
// supply HTTP/provider adapters, never a second feature implementation tree.
func TestPublishedModulesOwnImplementationOutsideHostFeatures(t *testing.T) {
	root := repoRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "..", "docs", "module-catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	var catalog struct {
		Backend []struct{ ID, Stage, ModuleID, Integration string }
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatal(err)
	}
	published := map[string]string{}
	for _, entry := range catalog.Backend {
		if entry.Stage != "published" {
			continue
		}
		published[entry.ID] = entry.ModuleID
		legacy := filepath.Join(root, "internal", "features", entry.ID)
		if _, err := os.Stat(legacy); !os.IsNotExist(err) {
			t.Errorf("published module %s retains a host feature tree: %v", entry.ID, err)
		}
		implementation := filepath.Join(root, "modules", entry.ModuleID, "go.mod")
		if _, err := os.Stat(implementation); err != nil {
			t.Errorf("published implementation missing: %s: %v", implementation, err)
		}
		adapter := filepath.Join(root, "internal", "adapters", "modules", entry.ID)
		if entry.Integration == "self-contained" {
			for _, retained := range []string{
				filepath.Join(root, "internal", "repository", entry.ID),
				filepath.Join(root, "internal", "service", entry.ID+"_service.go"),
			} {
				if _, err := os.Stat(retained); !os.IsNotExist(err) {
					t.Errorf("self-contained module retains host implementation: %s", retained)
				}
			}
			assets := filepath.Join(root, "assets", entry.ModuleID)
			if _, err := os.Stat(assets); !os.IsNotExist(err) {
				t.Errorf("self-contained module retains loose host assets: %s", assets)
			}

			if _, err := os.Stat(adapter); !os.IsNotExist(err) {
				t.Errorf("self-contained module retains adapter: %s", adapter)
			}
			continue
		}
		selected := false
		err := filepath.WalkDir(adapter, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, spec := range file.Imports {
				name, err := strconv.Unquote(spec.Path.Value)
				if err != nil {
					return err
				}
				expected := "github.com/mohuganda/moh-sso-dashboard/backend/modules/" + published[filepath.Base(adapter)]
				if name == expected || strings.HasPrefix(name, expected+"/") {
					selected = true
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if !selected {
			t.Errorf("%s does not delegate to its published module", adapter)
		}
	}
	if len(published) == 0 {
		t.Fatal("published catalogue unexpectedly empty")
	}
	err = filepath.WalkDir(filepath.Join(root, "internal"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			name, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			if strings.HasPrefix(name, modulePath) {
				feature := strings.Split(strings.TrimPrefix(name, modulePath), "/")[0]
				if _, published := published[feature]; published {
					t.Errorf("%s imports a retained published feature: %s", path, name)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
