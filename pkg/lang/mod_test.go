package lang_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aethiopicuschan/goimportsruler/pkg/lang"
	"github.com/stretchr/testify/assert"
)

func TestFindModuleRoot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(t *testing.T) (startPath string, wantRoot string)
		wantError bool
	}{
		{
			name: "find from module root directory",
			setup: func(t *testing.T) (string, string) {
				t.Helper()

				root := t.TempDir()
				err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/m\n\ngo 1.24\n"), 0o644)
				assert.NoError(t, err)
				return root, root
			},
		},
		{
			name: "find from nested directory",
			setup: func(t *testing.T) (string, string) {
				t.Helper()

				root := t.TempDir()
				err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/m\n\ngo 1.24\n"), 0o644)
				assert.NoError(t, err)

				nested := filepath.Join(root, "a", "b", "c")
				err = os.MkdirAll(nested, 0o755)
				assert.NoError(t, err)

				return nested, root
			},
		},
		{
			name: "find from file path under module",
			setup: func(t *testing.T) (string, string) {
				t.Helper()

				root := t.TempDir()
				err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/m\n\ngo 1.24\n"), 0o644)
				assert.NoError(t, err)

				dir := filepath.Join(root, "x", "y")
				err = os.MkdirAll(dir, 0o755)
				assert.NoError(t, err)

				f := filepath.Join(dir, "z.txt")
				err = os.WriteFile(f, []byte("dummy"), 0o644)
				assert.NoError(t, err)

				return f, root
			},
		},
		{
			name: "error when no go.mod exists upward",
			setup: func(t *testing.T) (string, string) {
				t.Helper()
				root := t.TempDir()
				return root, ""
			},
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			start, wantRoot := tc.setup(t)
			got, err := lang.FindModuleRoot(start)

			if tc.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, wantRoot, got)
		})
	}
}

func TestReadModuleNameFrom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		goMod     string
		want      string
		wantError bool
	}{
		{
			name:  "reads module path",
			goMod: "module example.com/m\n\ngo 1.24\n",
			want:  "example.com/m",
		},
		{
			name:      "invalid go.mod returns error",
			goMod:     "this is not a go.mod\n",
			wantError: true,
		},
		{
			name:  "no module directive returns empty",
			goMod: "go 1.24\n",
			want:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			err := os.WriteFile(filepath.Join(root, "go.mod"), []byte(tc.goMod), 0o644)
			assert.NoError(t, err)

			got, err := lang.ReadModuleNameFrom(root)

			if tc.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
