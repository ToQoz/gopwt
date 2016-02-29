package main

import (
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/gcimporter"
	"golang.org/x/tools/go/types"
	"os"
	"os/exec"
)

var typesInfo *types.Info

func getTypeInfo(pkgDir, importPath, tempGoSrcDir string, fset *token.FileSet, files []*ast.File) (*types.Info, error) {
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
	//       (e.g. github.com/ToQoz/gopwt/rewriter.go:201:29: cannot pass argument token.NewFileSet() (value of type *go/token.FileSet) to parameter of type *go/token.FileSet)

	// Install binaries
	deps, err := findDeps(importPath, tempGoSrcDir)
	if err != nil {
		return nil, err
	}
	if containsGoFile2(fset) {
		// install self
		deps = append(deps, ".")
	}
	if len(deps) > 0 {
		install := exec.Command("go", "install")
		install.Dir = pkgDir
		install.Stdout = os.Stdout
		install.Stderr = os.Stderr
		if *verbose {
			install.Args = append(install.Args, "-v")
		}
		install.Args = append(install.Args, deps...)
		if err := install.Run(); err != nil {
			return nil, err
		}
	}

	// Assume types from ast.Node
	typesConfig.Import = gcimporter.Import
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
	err = types.NewChecker(&typesConfig, fset, pkg, info).Files(files)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func determinantExprOfIsTypeConversion(e ast.Expr) ast.Expr {
	switch e.(type) {
	case *ast.ParenExpr:
		return determinantExprOfIsTypeConversion(e.(*ast.ParenExpr).X)
	case *ast.StarExpr:
		return determinantExprOfIsTypeConversion(e.(*ast.StarExpr).X)
	case *ast.CallExpr:
		return determinantExprOfIsTypeConversion(e.(*ast.CallExpr).Fun)
	case *ast.SelectorExpr:
		return e.(*ast.SelectorExpr).Sel
	default:
		return e
	}
}

func isTypeConversion(info *types.Info, e *ast.CallExpr) bool {
	if typesInfo == nil {
		return false
	}

	funcOrType := determinantExprOfIsTypeConversion(e)

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
