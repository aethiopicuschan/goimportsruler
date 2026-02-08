package lang

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/aethiopicuschan/goimportsruler/pkg/errs"
	"github.com/aethiopicuschan/goimportsruler/pkg/values"
	"golang.org/x/mod/modfile"
)

// FindModuleRoot walks up from startPath to find a directory containing go.mod.
// startPath can be a file or directory path.
func FindModuleRoot(startPath string) (root string, err error) {
	if startPath == "" {
		startPath = "."
	}

	abs, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	fi, err := os.Stat(abs)
	if err != nil {
		return "", err
	}

	dir := abs
	if !fi.IsDir() {
		dir = filepath.Dir(abs)
	}

	for {
		candidate := filepath.Join(dir, values.GO_MOD_FILE_NAME)
		if _, statErr := os.Stat(candidate); statErr == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("%w %s", errs.ErrReadFailed, values.GO_MOD_FILE_NAME)
}

// ReadModuleNameFrom reads <moduleRoot>/go.mod and returns the module path.
func ReadModuleNameFrom(moduleRoot string) (mn string, err error) {
	fn := filepath.Join(moduleRoot, values.GO_MOD_FILE_NAME)

	data, err := os.ReadFile(fn)
	if err != nil {
		err = fmt.Errorf("%w %s", errs.ErrReadFailed, fn)
		return
	}

	f, err := modfile.Parse(values.GO_MOD_FILE_NAME, data, nil)
	if err != nil {
		err = fmt.Errorf("%w %s", errs.ErrInvalidFile, fn)
		return
	}

	if f.Module == nil {
		return
	}

	mn = f.Module.Mod.Path
	return
}
