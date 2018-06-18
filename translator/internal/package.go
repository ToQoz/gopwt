package internal

import (
	"go/ast"
	"go/token"
	"go/types"
	"io"
	"os"
	"strings"
)

type GoFile struct {
	Path       string
	Original   *ast.File
	Normalized *ast.File
	Mode       os.FileMode
	Data       io.Reader
}

type PackageContext struct {
	Filepath   string
	Importpath string
	SrcDir     string

	Vendor    string
	HasVendor bool

	TypeInfo       *types.Info
	OriginalFset   *token.FileSet
	NormalizedFset *token.FileSet
	GoFiles        []*GoFile

	Error error
}

func NewPackageContext(fpath, importpath, srcDir string) *PackageContext {
	return &PackageContext{
		Filepath:       fpath,
		Importpath:     importpath,
		SrcDir:         srcDir,
		OriginalFset:   token.NewFileSet(),
		NormalizedFset: token.NewFileSet(),
	}
}

func (pkgCtx *PackageContext) Translate() error {
	pkgCtx.Vendor, pkgCtx.HasVendor = FindVendor(pkgCtx.Filepath, strings.Count(pkgCtx.Importpath, "/")+1)

	pkgCtx.CopyPackage()
	if pkgCtx.Error != nil {
		return pkgCtx.Error
	}

	pkgCtx.TypecheckPackage()
	if pkgCtx.Error != nil {
		return pkgCtx.Error
	}

	pkgCtx.RewritePackage()
	return pkgCtx.Error
}
