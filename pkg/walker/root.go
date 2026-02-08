package walker

import (
	"os"

	"github.com/aethiopicuschan/goimportsruler/pkg/errs"
	"github.com/aethiopicuschan/goimportsruler/pkg/values"
)

// IsRoot checks if the current execution context is the root of the module.
func IsRoot() (err error) {
	fi, err := os.Stat(values.GO_MOD_FILE_NAME)
	if err != nil {
		return errs.ErrNotRoot
	}
	if fi.IsDir() {
		return errs.ErrNotRoot
	}
	return
}
