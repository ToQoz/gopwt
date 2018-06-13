package internal

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/importer"
	"io"
	"path/filepath"
	"runtime"
	// "go/internal/gcimporter"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"strings"
)

func GetTypeInfo(vendor string, hasVendor bool, pkgDir, importPath, tempGoSrcDir string, fset *token.FileSet, files []*ast.File) (*types.Info, error) {
	deps, err := findDeps(importPath, tempGoSrcDir)
	if err != nil {
		return nil, err
	}

	installDeps := []string{}
	if hasVendor {
		for _, dep := range deps {
			if _, err := os.Stat(filepath.Join(vendor, dep)); err == nil {
				// ignore
				continue
			}
			installDeps = append(installDeps, dep)
		}
		pkgToVendor, err := filepath.Rel(pkgDir, vendor)
		if err != nil {
			return nil, err
		}
		p := filepath.Join(pkgToVendor, "...")
		if !strings.HasPrefix(p, ".") {
			p = "./" + p
		}
		installDeps = append(installDeps, p)
	} else {
		installDeps = deps
	}

	if IsBuildableFileSet(fset) {
		installDeps = append(installDeps, "./...")
	}

	if len(installDeps) > 0 {
		install := exec.Command("go", "install")
		install.Dir = pkgDir
		install.Stdout = os.Stdout
		b := []byte{}
		buf := bytes.NewBuffer(b)
		install.Stderr = buf
		if Verbose {
			install.Args = append(install.Args, "-v")
		}
		install.Args = append(install.Args, installDeps...)
		if err := install.Run(); err != nil {
			return nil, fmt.Errorf("[ERROR] go install %s\n\n%s", strings.Join(installDeps, " "), buf.String())
		}
	}

	if hasVendor {
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
				return os.MkdirAll(outpath, 0755)
			}

			if finfo.IsDir() {
				return os.MkdirAll(outpath, finfo.Mode())
			}

			out, err := os.OpenFile(outpath, os.O_RDWR|os.O_CREATE, finfo.Mode())
			if err != nil {
				return err
			}
			defer out.Close()
			in, err := os.Open(path)
			if err != nil {
				return err
			}
			defer in.Close()

			_, err = io.Copy(out, in)
			return err
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

	// WORKARROUND(xtest): if the tests is xtest like a `{pkg}_test`, check types only in `{pkg}_test`
	// NOTE: if your check types `package {pkg}_test` and `package {pkg}`, you'll get `package {pkg}_test; expected {pkg}`
	isXtest := false
	xtests := []*ast.File{}
	for _, f := range files {
		if strings.HasSuffix(f.Name.Name, "_test") {
			xtests = append(xtests, f)
		}
	}
	isXtest = len(xtests) > 0 && len(xtests) != len(files)

	if isXtest {
		err = types.NewChecker(&typesConfig, fset, pkg, info).Files(xtests)
	} else {
		err = types.NewChecker(&typesConfig, fset, pkg, info).Files(files)
	}
	if err != nil {
		return nil, err
	}

	return info, nil
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

	panic("unexpected error")
}
