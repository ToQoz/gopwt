package internal

import (
	"go/ast"
	"os"
	"path/filepath"
)

type Context struct {
	TranslatedassertImport *ast.Ident
	AssertImport           *ast.Ident
}

var (
	Testdata   = "testdata"
	TermWidth  = 0
	WorkingDir = ""
	Verbose    = false
)

func Translate(gopath string, importpath, fpath string) error {
	srcDir := filepath.Join(gopath, "src")
	err := os.MkdirAll(filepath.Join(srcDir, importpath), os.ModePerm)
	if err != nil {
		return err
	}

	return NewPackageContext(fpath, importpath, srcDir).Translate()
}
