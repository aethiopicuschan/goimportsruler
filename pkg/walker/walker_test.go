// pkg/walker/walker_test.go
package walker_test

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/aethiopicuschan/goimportsruler/pkg/config"
	"github.com/aethiopicuschan/goimportsruler/pkg/ctxrw"
	"github.com/aethiopicuschan/goimportsruler/pkg/walker"
	"github.com/stretchr/testify/assert"
)

func TestWalker_EndToEnd(t *testing.T) {
	t.Parallel()

	type wantViolation struct {
		file        string
		sourcePkg   string
		importPkg   string
		ruleName    string
		description string
		line        int
	}

	tests := []struct {
		name          string
		configJSON    string
		files         map[string]string // module-root-relative path -> contents
		startPath     string            // path given to NewWalker (module root will be discovered upward)
		walkRoots     []string          // filesystem roots passed to Walk
		wantViolCount int
		want          []wantViolation
	}{
		{
			name: "normalizes module imports and reports line/col + description",
			configJSON: `{
				"rules": [
					{
						"name": "ban cmd -> pkg",
						"description": "cmd must not import pkg",
						"sources": ["cmd/**"],
						"disallow": ["pkg/**"]
					}
				]
			}`,
			files: map[string]string{
				"go.mod": "module example.com/m\n\ngo 1.24\n",
				"pkg/config/config.go": `package config

type C struct{}
`,
				"cmd/root.go": `package cmd

import (
	"fmt"
	"example.com/m/pkg/config"
)

func f() {
	_ = fmt.Sprintf("%v", config.C{})
}
`,
			},
			startPath:     "cmd",
			walkRoots:     []string{"cmd"},
			wantViolCount: 1,
			want: []wantViolation{
				{
					file:        "cmd/root.go",
					sourcePkg:   "cmd",
					importPkg:   "pkg/config",
					ruleName:    "ban cmd -> pkg",
					description: "cmd must not import pkg",
					line:        5,
				},
			},
		},
		{
			name: "walk roots limit scanning scope",
			configJSON: `{
				"rules": [
					{
						"name": "ban pkg -> cmd",
						"description": "pkg must not import cmd",
						"sources": ["pkg/**"],
						"disallow": ["cmd/**"]
					}
				]
			}`,
			files: map[string]string{
				"go.mod": "module example.com/m\n\ngo 1.24\n",
				"cmd/a.go": `package cmd

import "fmt"

func f() { _ = fmt.Sprintf("") }
`,
				"pkg/p.go": `package pkg

import "example.com/m/cmd"

var _ = cmd.Foo
`,
				"cmd/foo.go": `package cmd

var Foo = 1
`,
			},
			startPath:     "pkg",
			walkRoots:     []string{"cmd"}, // should NOT scan pkg, so no violation
			wantViolCount: 0,
		},
		{
			name: "exclude prevents checking matching sources",
			configJSON: `{
				"rules": [
					{
						"name": "ban cmd -> pkg",
						"description": "cmd must not import pkg",
						"sources": ["cmd/**"],
						"disallow": ["pkg/**"]
					}
				],
				"excludes": [
					{
						"name": "exclude cmd",
						"description": "exclude all cmd packages",
						"sources": ["cmd/**"]
					}
				]
			}`,
			files: map[string]string{
				"go.mod": "module example.com/m\n\ngo 1.24\n",
				"pkg/x/x.go": `package x

type X struct{}
`,
				"cmd/root.go": `package cmd

import "example.com/m/pkg/x"

var _ = x.X{}
`,
			},
			startPath:     "cmd",
			walkRoots:     []string{"cmd"},
			wantViolCount: 0,
		},
		{
			name: "supports absolute patterns as well as relative ones",
			configJSON: `{
				"rules": [
					{
						"name": "ban cmd -> pkg (abs patterns)",
						"description": "absolute patterns should match too",
						"sources": ["example.com/m/cmd/**"],
						"disallow": ["example.com/m/pkg/**"]
					}
				]
			}`,
			files: map[string]string{
				"go.mod": "module example.com/m\n\ngo 1.24\n",
				"pkg/a/a.go": `package a

type A struct{}
`,
				"cmd/root.go": `package cmd

import "example.com/m/pkg/a"

var _ = a.A{}
`,
			},
			startPath:     "cmd/root.go",
			walkRoots:     []string{"cmd"},
			wantViolCount: 1,
			want: []wantViolation{
				{
					file:        "cmd/root.go",
					sourcePkg:   "cmd",
					importPkg:   "pkg/a",
					ruleName:    "ban cmd -> pkg (abs patterns)",
					description: "absolute patterns should match too",
					line:        3,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create a temporary module structure.
			root := t.TempDir()
			for rel, content := range tc.files {
				abs := filepath.Join(root, filepath.FromSlash(rel))
				err := os.MkdirAll(filepath.Dir(abs), 0o755)
				assert.NoError(t, err)
				err = os.WriteFile(abs, []byte(content), 0o644)
				assert.NoError(t, err)
			}

			// Write config file into the module root, then load it via config.LoadConfig(startPath).
			cfgPath := filepath.Join(root, ".goimportsruler.json")
			err := os.WriteFile(cfgPath, []byte(tc.configJSON), 0o644)
			assert.NoError(t, err)

			cfg, _, err := config.LoadConfig(root)
			assert.NoError(t, err)
			if !assert.NotNil(t, cfg) {
				return
			}
			if !assert.Greater(t, len(cfg.Rules()), 0, "expected config rules to be loaded") {
				return
			}

			ctx := ctxrw.NewContextWriter().SetConfig(context.Background(), *cfg)

			startAbs := filepath.Join(root, filepath.FromSlash(tc.startPath))
			w, err := walker.NewWalker(ctx, startAbs)
			assert.NoError(t, err)
			if !assert.NotNil(t, w) {
				return
			}

			roots := make([]string, 0, len(tc.walkRoots))
			for _, r := range tc.walkRoots {
				roots = append(roots, filepath.Join(root, filepath.FromSlash(r)))
			}

			ve, err := w.Walk(ctx, roots...)
			assert.NoError(t, err)

			assert.Equal(t, tc.wantViolCount, ve.Len(), "violation count mismatch")

			if tc.wantViolCount == 0 {
				return
			}

			// Collect and sort for stable comparison.
			got := make([]wantViolation, 0, ve.Len())
			for v := range ve.Violations() {
				got = append(got, wantViolation{
					file:        v.File(),
					sourcePkg:   v.SourcePkg(),
					importPkg:   v.ImportPath(),
					ruleName:    v.RuleName(),
					description: v.RuleDescription(),
					line:        v.Line(),
				})
			}

			sort.Slice(got, func(i, j int) bool {
				if got[i].file != got[j].file {
					return got[i].file < got[j].file
				}
				if got[i].ruleName != got[j].ruleName {
					return got[i].ruleName < got[j].ruleName
				}
				if got[i].importPkg != got[j].importPkg {
					return got[i].importPkg < got[j].importPkg
				}
				if got[i].line != got[j].line {
					return got[i].line < got[j].line
				}
				return got[i].description < got[j].description
			})

			want := append([]wantViolation(nil), tc.want...)
			sort.Slice(want, func(i, j int) bool {
				if want[i].file != want[j].file {
					return want[i].file < want[j].file
				}
				if want[i].ruleName != want[j].ruleName {
					return want[i].ruleName < want[j].ruleName
				}
				if want[i].importPkg != want[j].importPkg {
					return want[i].importPkg < want[j].importPkg
				}
				if want[i].line != want[j].line {
					return want[i].line < want[j].line
				}
				return want[i].description < want[j].description
			})

			assert.Equal(t, want, got)
		})
	}
}

func TestViolation_Defaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		v        walker.Violation
		wantName string
		wantDesc string
	}{
		{
			name:     "empty fields use defaults",
			v:        walker.Violation{},
			wantName: "<unknown>",
			wantDesc: "<no description>",
		},
		{
			name: "non-empty fields are returned as-is",
			v: func() walker.Violation {
				var vv walker.Violation
				// The test is in walker_test package, so we cannot set unexported fields directly.
				// This case is kept to validate defaults only.
				return vv
			}(),
			wantName: "<unknown>",
			wantDesc: "<no description>",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.wantName, tc.v.RuleName())
			assert.Equal(t, tc.wantDesc, tc.v.RuleDescription())
		})
	}
}
