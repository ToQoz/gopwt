package internal

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"path/filepath"
	"runtime"
	// "go/internal/gcimporter"

	"go/types"
	"os"
	"os/exec"
	"strings"
)

func GetTypeInfo(pkgCtx *PackageContext, pkgDir, importPath, tempGoSrcDir string) (*types.Info, error) {
	buildPackage, err := build.Import(importPath, tempGoSrcDir, build.AllowBinary)
	if err != nil {
		return nil, err
	}

	deps, err := FindDeps(buildPackage)
	if err != nil {
		return nil, err
	}

	for i, dep := range deps {
		// NOTE: rewrite $dep to $importpath/vendor/$dep if $dep is found in vendor dir
		if pkgCtx.HasVendor {
			if _, err := os.Stat(filepath.Join(pkgCtx.Vendor, dep)); err == nil {
				pkgToVendor, err := filepath.Rel(pkgDir, pkgCtx.Vendor)
				if err != nil {
					return nil, err
				}
				dep = filepath.Join(pkgToVendor, dep)
				if !strings.HasPrefix(dep, ".") {
					dep = "./" + dep
				}
			}
		}

		deps[i] = dep
	}

	if IsBuildableFileSet(pkgCtx.NormalizedFset) {
		deps = append(deps, ".")
	}

	if len(deps) > 0 {
		install := exec.Command("go", "install")
		install.Dir = pkgDir
		install.Stdout = os.Stdout
		b := []byte{}
		buf := bytes.NewBuffer(b)
		install.Stderr = buf
		if Verbose {
			install.Args = append(install.Args, "-v")
		}
		install.Args = append(install.Args, deps...)
		if err := install.Run(); err != nil {
			return nil, fmt.Errorf("[ERROR] go install %s\n\n%s", strings.Join(deps, " "), buf.String())
		}
	}

	if pkgCtx.HasVendor {
		// NOTE: move pkg/$importpath/vendor/$v-importpath to pkg/$v-importpath
		pd := filepath.Join(os.Getenv("GOPATH"), "pkg", runtime.GOOS+"_"+runtime.GOARCH)
		err := filepath.Walk(pd, func(path string, finfo os.FileInfo, err error) error {
			if path == pd {
				return nil
			}

			vimportpath, found := RetrieveImportpathFromVendorDir(path)
			if !found {
				return nil
			}

			outpath := filepath.Join(pd, vimportpath)
			if finfo.IsDir() {
				return os.MkdirAll(outpath, finfo.Mode())
			}
			out, err := os.OpenFile(outpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, finfo.Mode())
			if err != nil {
				return err
			}
			defer out.Close()
			return CopyFile(path, out)
		})
		if err != nil {
			return nil, err
		}
	}

	// Assume types from ast.Node
	typesConfig := types.Config{}
	typesConfig.Importer = importer.Default()
	pkg := types.NewPackage(importPath, "")
	info := &types.Info{
		Types:      map[ast.Expr]types.TypeAndValue{},
		Defs:       map[*ast.Ident]types.Object{},
		Uses:       map[*ast.Ident]types.Object{},
		Implicits:  map[ast.Node]types.Object{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
		Scopes:     map[ast.Node]*types.Scope{},
		InitOrder:  []*types.Initializer{},
	}

	// WORKARROUND(xtest): if the tests is xtest like `{pkg}_test`, check types only in `{pkg}_test`
	// NOTE: if you check types both `package {pkg}_test` and `package {pkg}`, you'll get `package {pkg}_test; expected {pkg}`
	isXtest := false
	xtests := []*ast.File{}
	for _, f := range pkgCtx.NormalizedFiles {
		if strings.HasSuffix(f.Name.Name, "_test") {
			xtests = append(xtests, f)
		}
	}
	isXtest = len(xtests) > 0 && len(xtests) != len(pkgCtx.NormalizedFiles)

	if isXtest {
		err = types.NewChecker(&typesConfig, pkgCtx.NormalizedFset, pkg, info).Files(xtests)
	} else {
		err = types.NewChecker(&typesConfig, pkgCtx.NormalizedFset, pkg, info).Files(pkgCtx.NormalizedFiles)
	}
	return info, err
}

func DeterminantExprOfIsTypeConversion(e ast.Expr) ast.Expr {
	switch e.(type) {
	case *ast.ParenExpr:
		return DeterminantExprOfIsTypeConversion(e.(*ast.ParenExpr).X)
	case *ast.StarExpr:
		return DeterminantExprOfIsTypeConversion(e.(*ast.StarExpr).X)
	case *ast.CallExpr:
		return DeterminantExprOfIsTypeConversion(e.(*ast.CallExpr).Fun)
	case *ast.SelectorExpr:
		return e.(*ast.SelectorExpr).Sel
	default:
		return e
	}
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
