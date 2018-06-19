package internal

import (
	"go/ast"
	"log"
	"os"
	"path/filepath"
)

type Context struct {
	TranslatedassertImport *ast.Ident
	AssertImport           *ast.Ident
}

var (
	cacheDir   = filepath.Join(os.Getenv(homeEnv), ".gopwtcache")
	Testdata   = "testdata"
	TermWidth  = 0
	WorkingDir = ""
	Verbose    = false
)

func Translate(gopath string, importpath, fpath string) error {
	if debugLog {
		log.Printf("path=%s", fpath)
		log.Printf("gopath=%s", gopath)
		log.Printf("importpath=%s", importpath)
	}

	srcDir := filepath.Join(gopath, "src")
	err := os.MkdirAll(filepath.Join(srcDir, importpath), os.ModePerm)
	if err != nil {
		return err
	}

	return NewPackageContext(fpath, importpath, srcDir).Translate()
}
