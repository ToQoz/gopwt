package internal

import (
	"go/ast"
	"strings"
	// "go/internal/gcimporter"

	"go/types"

	"golang.org/x/tools/go/loader"
)

func (pkgCtx *PackageContext) TypecheckPackage() {
	c := loader.Config{}
	var files []*ast.File
	var xtests []*ast.File
	for _, f := range pkgCtx.GoFiles {
		files = append(files, f.Normalized)
		if strings.HasSuffix(f.Normalized.Name.Name, "_test") {
			xtests = append(xtests, f.Normalized)
		}
	}

	isXtest := len(xtests) > 0 && len(xtests) != len(files)

	c.Fset = pkgCtx.NormalizedFset
	if isXtest {
		c.CreateFromFiles(pkgCtx.Importpath, xtests...)
	} else {
		c.CreateFromFiles(pkgCtx.Importpath, files...)
	}
	pkg, err := c.Load()
	if err != nil {
		pkgCtx.Error = err
		return
	}
	pkgCtx.TypeInfo = &pkg.Created[0].Info
	return
}

func IsTypeConversion(info *types.Info, e *ast.CallExpr) bool {
	if info == nil {
		return false
	}

	funcOrType := DeterminantExprOfIsTypeConversion(e)

	switch funcOrType.(type) {
	case *ast.ChanType, *ast.FuncType, *ast.MapType, *ast.ArrayType, *ast.StructType, *ast.InterfaceType:
		return true
	case *ast.Ident:
		id := funcOrType.(*ast.Ident)

		if t, ok := info.Types[id]; ok {
			return t.IsType()
		}

		if o := info.ObjectOf(id); o != nil {
			switch o.(type) {
			case *types.TypeName:
				return true
			default:
				return false
			}
		}
	}

	return false
}
