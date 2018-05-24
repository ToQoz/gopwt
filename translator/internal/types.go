package internal

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func GetTypeInfo(pkgDir, importPath, tempGoSrcDir string, fset *token.FileSet, files []*ast.File) (*types.Info, error) {
	typesConfig := types.Config{}

	// Install binaries
	buildPkg, err := build.Default.Import(importPath, "", build.AllowBinary)
	if err != nil {
		return nil, err
	}
	deps, err := findDeps(buildPkg, importPath)
	if err != nil {
		return nil, err
	}

	installDeps := []string{}
	if IsBuildableFileSet(fset) {
		// install self
		old, err := isOldBinary(buildPkg)
		if err != nil {
			return nil, err
		}
		if old {
			installDeps = append(installDeps, importPath)
		}
	}

	for _, d := range deps {
		pkg, err := build.Default.Import(d, "", build.FindOnly)
		if err != nil {
			return nil, err
		}
		old, err := isOldBinary(pkg)
		if err != nil {
			return nil, err
		}
		if old {
			installDeps = append(installDeps, d)
		}
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
			return nil, fmt.Errorf("[ERROR] go install %s\n\n%s", strings.Join(deps, " "), buf.String())
		}
	}

	// Assume types from ast.Node
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

func isOldBinary(pkg *build.Package) (bool, error) {
	ps, err := os.Stat(pkg.PkgObj)
	if err != nil {
		return true, nil
	}
	pt := ps.ModTime()
	fs, err := ioutil.ReadDir(pkg.Dir)
	if err != nil {
		return false, err
	}
	for _, f := range fs {
		if f.ModTime().After(pt) {
			return true, nil
		}
	}
	return false, nil
}
