package lang

import (
	"go/parser"
	"go/token"
	"strconv"
)

type Import struct {
	path string
	aka  string
	pos  token.Position
}

func (i *Import) Path() string {
	return i.path
}

func (i *Import) Aka() string {
	return i.aka
}

func (i *Import) Pos() token.Position {
	return i.pos
}

// GetImportPath parses the Go source code provided in src and returns a slice of Import structs,
// each containing the import path and its alias (if any).
func GetImportPath(src string) (imports []Import, err error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return
	}

	for _, imp := range file.Imports {
		var path string
		path, err = strconv.Unquote(imp.Path.Value)
		if err != nil {
			return
		}

		aka := ""
		if imp.Name != nil {
			aka = imp.Name.Name
		}

		pos := fset.Position(imp.Pos())
		imports = append(imports, Import{path: path, aka: aka, pos: pos})
	}

	return
}
