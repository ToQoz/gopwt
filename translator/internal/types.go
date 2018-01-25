package internal

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/importer"
	// "go/internal/gcimporter"
	"go/token"
	"go/types"
	"os"
	"os/exec"
	"strings"
)

func GetTypeInfo(pkgDir, importPath, tempGoSrcDir string, fset *token.FileSet, files []*ast.File) (*types.Info, error) {
	typesConfig := types.Config{}
	// I gave up to loading pkg from source.(by using "code.google.com/p/go.tools/go/loader")
	// 	typesConfig.Import = func(imports map[string]*types.Package, path string) (*types.Package, error) {
	// 		// Import from source if fail to import from binary
	// 		pkg, err := gcimporter.Import(imports, path)
	// 		if err == nil {
	// 			return pkg, nil
	// 		}
	//
	// 		lConfig := loader.Config{}
	// 		lConfig.TypeChecker = typesConfig
	// 		lConfig.Build = &build.Default
	// 		lConfig.SourceImports = true
	// 		lConfig.Import(path)
	//
	// 		prog, err := lConfig.Load()
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		fmt.Println(prog.Imported[path].Types)
	// 		return prog.Imported[path].Pkg, nil
	// 	}
	// Because
	//   - it is slow.
	//   - i met strange errors.
	//       (e.g. github.com/ToQoz/gopwt/translator.go:201:29: cannot pass argument token.NewFileSet() (value of type *go/token.FileSet) to parameter of type *go/token.FileSet)

	// Install binaries
	deps, err := findDeps(importPath, tempGoSrcDir)
	if err != nil {
		return nil, err
	}
	if IsBuildableFileSet(fset) {
		// install self
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

	// Assume types from ast.Node
	// typesConfig.Import = gcimporter.Import
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
