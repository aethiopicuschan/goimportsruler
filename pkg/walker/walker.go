// pkg/walker/walker.go
package walker

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/aethiopicuschan/goimportsruler/pkg/config"
	"github.com/aethiopicuschan/goimportsruler/pkg/ctxrw"
	"github.com/aethiopicuschan/goimportsruler/pkg/lang"
)

type Walker struct {
	rootModule string
	moduleRoot string // absolute filesystem path to the directory containing go.mod
	ctx        context.Context
}

// NewWalker creates a Walker by locating go.mod from startPath and reading its module path.
// startPath can be ".", "./foobar", or any path inside the target module.
func NewWalker(ctx context.Context, startPath string) (w *Walker, err error) {
	mr, err := lang.FindModuleRoot(startPath)
	if err != nil {
		return nil, err
	}
	mn, err := lang.ReadModuleNameFrom(mr)
	if err != nil {
		return nil, err
	}

	w = &Walker{
		rootModule: mn,
		moduleRoot: mr,
		ctx:        ctx,
	}
	return
}

// Walk walks through Go files under the given roots, checks import rules,
// and returns a ViolationsError if any rule is violated.
//
// - roots are filesystem paths (e.g. ".", "./foobar", "internal").
// - Import path normalization is computed relative to the module root directory.
// - If ctx is nil, the Walker's context (set in NewWalker) is used.
func (w *Walker) Walk(ctx context.Context, roots ...string) (ve ViolationsError, err error) {
	if ctx == nil {
		ctx = w.ctx
	}
	if len(roots) == 0 {
		roots = []string{"."}
	}

	cfg, err := ctxrw.NewContextReader().GetConfig(ctx)
	if err != nil {
		return
	}

	var violations []Violation

	for _, root := range roots {
		absRoot, err2 := filepath.Abs(root)
		if err2 != nil {
			return ve, err2
		}

		walkErr := filepath.WalkDir(absRoot, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			// Skip common VCS directories early for speed.
			if d.IsDir() {
				base := filepath.Base(path)
				switch base {
				case ".git", ".hg", ".svn":
					return fs.SkipDir
				}
				return nil
			}

			// Only Go files.
			if !strings.HasSuffix(path, ".go") {
				return nil
			}

			// Build both absolute (module-prefixed) and relative (module-less) source package paths.
			// Note: The relative import path is computed from the module root, not from the walk root.
			srcAbs, srcRel := w.fileToImportPaths(path)

			// If excluded, skip checks entirely for this file's package.
			if isExcluded(cfg, srcAbs, srcRel) {
				return nil
			}

			b, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}

			imps, parseErr := lang.GetImportPath(string(b))
			if parseErr != nil {
				// If parsing fails, treat it as an error because imports cannot be determined.
				return parseErr
			}

			for _, imp := range imps {
				impAbs, impRel := w.importToImportPaths(imp.Path())

				for _, rule := range cfg.Rules() {
					// Apply rules only when the source matches (support both absolute and relative forms).
					if !anyMatch2(rule.Sources(), srcAbs, srcRel) {
						continue
					}

					// If the import matches any allowed pattern, skip to the next rule without checking disallowed patterns.
					if len(rule.Excludes()) > 0 {
						if anyMatch2(rule.Excludes(), srcAbs, srcRel) {
							continue // Allowed import, skip to next rule.
						}
					}

					// A violation occurs if the import matches any disallowed pattern (absolute or relative).
					if anyMatch2(rule.Disallow(), impAbs, impRel) {
						pos := imp.Pos()
						fileRel := w.fileToDisplayPath(path)

						violations = append(violations, Violation{
							file:        fileRel,
							sourcePkg:   srcRel, // Use relative paths in output for readability.
							importPath:  impRel, // Use relative import path if it's in the same module.
							ruleName:    rule.Name(),
							description: rule.Description(),
							line:        pos.Line,
							column:      pos.Column,
						})
					}
				}
			}

			return nil
		})

		if walkErr != nil {
			return ve, walkErr
		}
	}

	if len(violations) > 0 {
		ve = ViolationsError{violations: violations}
		return
	}

	return
}

// fileToDisplayPath returns a module-root-relative file path for output.
func (w *Walker) fileToDisplayPath(absFilePath string) string {
	rel, err := filepath.Rel(w.moduleRoot, absFilePath)
	if err != nil {
		return filepath.ToSlash(absFilePath)
	}
	return filepath.ToSlash(rel)
}

// fileToImportPaths returns (abs, rel) import paths for the file's directory.
// - abs: "<module>/<relDir>" or "<module>" for module root
// - rel: "<relDir>" or "." for module root
//
// The relative directory is computed from the module root directory.
func (w *Walker) fileToImportPaths(absFilePath string) (abs string, rel string) {
	dir := filepath.Dir(absFilePath)

	r, err := filepath.Rel(w.moduleRoot, dir)
	if err != nil {
		// Fall back conservatively to module root.
		return w.rootModule, "."
	}

	r = filepath.ToSlash(r)
	if r == "." {
		return w.rootModule, "."
	}
	return w.rootModule + "/" + r, r
}

// importToImportPaths returns (abs, rel) import paths for an import string.
//   - abs: the original import path
//   - rel: module-less path if it belongs to the current module, otherwise the original import path.
//     Also returns "." when the import is exactly the module root.
func (w *Walker) importToImportPaths(importPath string) (abs string, rel string) {
	abs = importPath
	rel = importPath

	if importPath == w.rootModule {
		return abs, "."
	}

	prefix := w.rootModule + "/"
	if strings.HasPrefix(importPath, prefix) {
		rel = strings.TrimPrefix(importPath, prefix)
		if rel == "" {
			rel = "."
		}
	}
	return
}

func isExcluded(cfg config.Config, srcAbs, srcRel string) bool {
	return anyMatch2(cfg.Excludes(), srcAbs, srcRel)
}

// anyMatch2 matches patterns against both absolute and relative targets.
// This allows config to be written as either "cmd/**" or "<module>/cmd/**".
func anyMatch2(patterns []config.Package, absTarget, relTarget string) bool {
	for _, p := range patterns {
		pat := p.String()
		if globMatch(pat, absTarget) || globMatch(pat, relTarget) {
			return true
		}
	}
	return false
}

// globMatch supports:
// - *  : within one segment (no '/')
// - ** : zero or more segments (can cross '/')
func globMatch(pattern, target string) bool {
	p := strings.Trim(filepath.ToSlash(pattern), "/")
	t := strings.Trim(filepath.ToSlash(target), "/")

	if p == "**" {
		return true
	}

	pSeg := splitKeepEmpty(p)
	tSeg := splitKeepEmpty(t)

	var rec func(i, j int) bool
	rec = func(i, j int) bool {
		if i == len(pSeg) && j == len(tSeg) {
			return true
		}
		if i == len(pSeg) {
			return false
		}

		// "**" matches zero or more segments.
		if pSeg[i] == "**" {
			// Consume zero segments.
			if rec(i+1, j) {
				return true
			}
			// Consume one segment.
			if j < len(tSeg) && rec(i, j+1) {
				return true
			}
			return false
		}

		if j == len(tSeg) {
			return false
		}

		ok, err := filepath.Match(pSeg[i], tSeg[j])
		if err != nil {
			// Invalid pattern segment: treat as non-match.
			return false
		}
		if !ok {
			return false
		}
		return rec(i+1, j+1)
	}

	return rec(0, 0)
}

func splitKeepEmpty(s string) []string {
	if s == "" {
		return []string{""}
	}
	return strings.Split(s, "/")
}
