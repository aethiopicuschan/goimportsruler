package lang_test

import (
	"testing"

	"github.com/aethiopicuschan/goimportsruler/pkg/lang"
	"github.com/stretchr/testify/assert"
)

func TestGetImportPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  string
		want []struct {
			path string
			aka  string
			line int
		}
		wantError bool
	}{
		{
			name: "single import",
			src: `package p

import "fmt"

func f() { _ = fmt.Sprintf("") }
`,
			want: []struct {
				path string
				aka  string
				line int
			}{
				{path: "fmt", aka: "", line: 3},
			},
		},
		{
			name: "import block with aliases",
			src: `package p

import (
	"fmt"
	alias "net/http"
	_ "embed"
	. "math"
)

func f() { _, _, _ = fmt.Sprintf, alias.MethodGet, Pi }
`,
			want: []struct {
				path string
				aka  string
				line int
			}{
				{path: "fmt", aka: "", line: 4},
				{path: "net/http", aka: "alias", line: 5},
				{path: "embed", aka: "_", line: 6},
				{path: "math", aka: ".", line: 7},
			},
		},
		{
			name:      "invalid source returns error",
			src:       `package p import "fmt"`,
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			imps, err := lang.GetImportPath(tc.src)

			if tc.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if !assert.Len(t, imps, len(tc.want)) {
				return
			}

			for i := range tc.want {
				assert.Equal(t, tc.want[i].path, imps[i].Path(), "path[%d]", i)
				assert.Equal(t, tc.want[i].aka, imps[i].Aka(), "aka[%d]", i)
				assert.Equal(t, tc.want[i].line, imps[i].Pos().Line, "line[%d]", i)
			}
		})
	}
}
