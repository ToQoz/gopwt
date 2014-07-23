package main

import (
	"bytes"
	"code.google.com/p/go.tools/go/gcimporter"
	"code.google.com/p/go.tools/go/types"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	translatedassertImportIdent = &ast.Ident{Name: "translatedassert"}
	assertImportIdent           = &ast.Ident{Name: "assert"}
	errAssertImportNotFound     = errors.New("github.com/ToQoz/gopwt/assert is not found in imports")
	builtinFuncs                = []string{
		"append",
		"cap",
		"close",
		"complex",
		"copy",
		"delete",
		"imag",
		"len",
		"make",
		"new",
		"panic",
		"print",
		"println",
		"real",
		"recover",
	}
	typesInfo *types.Info
)

func rewritePackage(pkgDir, importPath string, tempGoSrcDir string) error {
	// Copy to tempdir
	err := copyPackage(pkgDir, importPath, tempGoSrcDir)
	if err != nil {
		return err
	}

	// Collect fset(*token.FileSet) and files([]*ast.File)
	fset := token.NewFileSet()
	files := []*ast.File{}
	root := filepath.Join(tempGoSrcDir, importPath)
	err = filepath.Walk(root, func(path string, fInfo os.FileInfo, err error) error {
		if fInfo.IsDir() {
			if path == root {
				return nil
			}

			return filepath.SkipDir
		}

		if !isGoFile2(path) {
			return nil
		}

		a, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}

		files = append(files, a)
		return nil
	})
	if err != nil {
		return err
	}

	typesInfo = getTypeInfo(importPath, fset, files)

	// Rewrite files
	for _, f := range files {
		path := fset.File(f.Package).Name()

		if !isTestGoFile(path) {
			continue
		}

		assertImportIdent = &ast.Ident{Name: "assert"}
		assertImport := getAssertImport(f)
		if assertImport == nil {
			continue
		}
		assertImport.Path.Value = `"github.com/ToQoz/gopwt/translatedassert"`
		if assertImport.Name != nil {
			assertImportIdent = assertImport.Name
			assertImport.Name = translatedassertImportIdent
		}

		fi, err := os.Stat(path)
		if err != nil {
			return err
		}

		out, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, fi.Mode())
		defer out.Close()
		err = rewriteFile(fset, f, out)
		if err != nil {
			return err
		}
	}

	return nil
}

func copyPackage(pkgDir, importPath string, tempGoSrcDir string) error {
	err := filepath.Walk(pkgDir, func(path string, fInfo os.FileInfo, err error) error {
		if fInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil
		}

		pathFromImportDir, err := filepath.Rel(pkgDir, path)
		if err != nil {
			return err
		}
		outPath := filepath.Join(tempGoSrcDir, importPath, pathFromImportDir)

		if fInfo.IsDir() {
			if path == pkgDir {
				return nil
			}

			// copy all files in <pkgDir>/testdata/**/*
			if strings.Split(pathFromImportDir, string(filepath.Separator))[0] == "testdata" {
				di, err := os.Stat(filepath.Dir(path))
				if err != nil {
					return err
				}
				err = os.Mkdir(filepath.Dir(outPath), di.Mode())
				if err != nil {
					return err
				}
				return nil
			}

			return filepath.SkipDir
		}

		out, err := os.OpenFile(outPath, os.O_RDWR|os.O_CREATE, fInfo.Mode())
		if err != nil {
			return err
		}
		defer out.Close()

		if !isTestGoFile(path) {
			return copyFile(path, out)
		}

		a, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}

		assertImportIdent = &ast.Ident{Name: "assert"}
		assertImport := getAssertImport(a)
		if assertImport == nil {
			return copyFile(path, out)
		}
		if assertImport.Name != nil {
			assertImportIdent = assertImport.Name
		}

		// Format and copy file
		//   - 1. replace rawStringLit to stringLit
		//   - 2. replace multiline compositLit to singleline one.
		//   - 3. copy
		ast.Inspect(a, func(n ast.Node) bool {
			if _, ok := n.(*ast.CallExpr); !ok {
				return true
			}

			c := n.(*ast.CallExpr)

			if _, ok := c.Fun.(*ast.FuncLit); ok {
				return true
			}

			// skip inspecting children in n
			if !isAssert(assertImportIdent, c) {
				return false
			}

			// 1. replace rawStringLit to stringLit
			replaceAllRawStringLitByStringLit(c)
			return true
		})

		// 2. replace multiline-compositLit by singleline-one by passing empty token.FileSet
		// 3. copy
		err = printer.Fprint(out, token.NewFileSet(), a)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func copyFile(path string, out io.Writer) error {
	filedata, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	_, err = out.Write(filedata)
	if err != nil {
		return err
	}

	return nil
}

func rewriteFile(fset *token.FileSet, a *ast.File, out io.Writer) error {
	ast.Inspect(a, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.CallExpr:
			n := n.(*ast.CallExpr)

			if _, ok := n.Fun.(*ast.FuncLit); ok {
				return true
			}

			// skip inspecting children in n
			if !isAssert(assertImportIdent, n) {
				return false
			}

			file := fset.File(n.Pos())
			filename := file.Name()
			line := file.Line(n.Pos())

			b := []byte{}
			buf := bytes.NewBuffer(b)
			// This printing **must** success.(valid expr -> code)
			// So call panic on failing.
			err := printer.Fprint(buf, token.NewFileSet(), n)
			if err != nil {
				panic(err)
			}

			n.Args = append(n.Args, createRawStringLit("FAIL"))
			n.Args = append(n.Args, createRawStringLit(filename))
			n.Args = append(n.Args, &ast.BasicLit{Value: strconv.Itoa(line), Kind: token.INT})
			n.Args = append(n.Args, createRawStringLit(buf.String()))
			n.Args = append(n.Args, &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(termw)})
			n.Args = append(n.Args, createPosValuePairExpr(extractPrintExprs(filename, line, n.Pos()-1, n, n.Args[1]))...)
			n.Fun.(*ast.SelectorExpr).X = &ast.Ident{Name: "translatedassert"}
			return false
		}

		return true
	})

	err := printer.Fprint(out, fset, a)
	if err != nil {
		return err
	}
	return nil
}

type printExpr struct {
	Pos  int
	Expr ast.Expr
}

func newPrintExpr(pos token.Pos, e ast.Expr) printExpr {
	return printExpr{Pos: int(pos), Expr: e}
}

func extractPrintExprs(filename string, line int, offset token.Pos, parent ast.Expr, n ast.Expr) []printExpr {
	ps := []printExpr{}

	switch n.(type) {
	case *ast.BasicLit:
		n := n.(*ast.BasicLit)
		if n.Kind == token.STRING {
			if len(strings.Split(n.Value, "\\n")) > 1 {
				ps = append(ps, newPrintExpr(n.Pos()-offset, n))
			}
		}
	case *ast.CompositeLit:
		n := n.(*ast.CompositeLit)

		for _, elt := range n.Elts {
			ps = append(ps, extractPrintExprs(filename, line, offset, n, elt)...)
		}
	case *ast.KeyValueExpr:
		n := n.(*ast.KeyValueExpr)

		if isMapType(parent) {
			ps = append(ps, extractPrintExprs(filename, line, offset, n, n.Key)...)
		}

		ps = append(ps, extractPrintExprs(filename, line, offset, n, n.Value)...)
	case *ast.Ident:
		ps = append(ps, newPrintExpr(n.Pos()-offset, n))
	case *ast.ParenExpr:
		n := n.(*ast.ParenExpr)
		ps = append(ps, extractPrintExprs(filename, line, offset, n, n.X)...)
	case *ast.StarExpr:
		n := n.(*ast.StarExpr)
		ps = append(ps, newPrintExpr(n.Pos()-offset, n))
		ps = append(ps, extractPrintExprs(filename, line, offset, n, n.X)...)
	case *ast.UnaryExpr:
		n := n.(*ast.UnaryExpr)
		x := extractPrintExprs(filename, line, offset, n, n.X)

		n.X = &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   translatedassertImportIdent,
				Sel: &ast.Ident{Name: "RVBool"},
			},
			Args: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X:   translatedassertImportIdent,
						Sel: &ast.Ident{Name: "RVOf"},
					},
					Args: []ast.Expr{n.X},
				},
			},
		}

		ps = append(ps, newPrintExpr(n.Pos()-offset, n))
		ps = append(ps, x...)
	case *ast.BinaryExpr:
		n := n.(*ast.BinaryExpr)

		var x, y []printExpr

		x = extractPrintExprs(filename, line, offset, n, n.X)
		y = extractPrintExprs(filename, line, offset, n, n.Y)

		newExpr := createUntypedExprFromBinaryExpr(n)
		if newExpr != n {
			replaceBinaryExpr(parent, n, newExpr)
		}

		ps = append(ps, x...)
		ps = append(ps, newPrintExpr(n.OpPos-offset, newExpr))
		ps = append(ps, y...)
	case *ast.IndexExpr:
		n := n.(*ast.IndexExpr)
		ps = append(ps, extractPrintExprs(filename, line, offset, n, n.X)...)
		ps = append(ps, extractPrintExprs(filename, line, offset, n, n.Index)...)
	case *ast.SelectorExpr:
		n := n.(*ast.SelectorExpr)
		ps = append(ps, extractPrintExprs(filename, line, offset, n, n.X)...)
		ps = append(ps, newPrintExpr(n.Sel.Pos()-offset, n))
	case *ast.CallExpr:
		n := n.(*ast.CallExpr)
		if isBuiltinFunc(n) { // don't memorize buildin methods
			newExpr := createUntypedCallExprFromBuiltinCallExpr(n)

			ps = append(ps, newPrintExpr(n.Pos()-offset, newExpr))
			for _, arg := range n.Args {
				ps = append(ps, extractPrintExprs(filename, line, offset, n, arg)...)
			}

			*n = *newExpr
		} else if isTypeConversion(typesInfo, n) {
			// T(v) can take only one argument.
			if len(n.Args) > 1 {
				panic("too many arguments for type conversion")
			} else if len(n.Args) < 1 {
				panic("missing argument for type conversion")
			}

			argsPrints := extractPrintExprs(filename, line, offset, n, n.Args[0])

			newExpr := &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   translatedassertImportIdent,
					Sel: &ast.Ident{Name: "RVInterface"},
				},
				Args: []ast.Expr{
					&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X: &ast.CallExpr{
								Fun: &ast.SelectorExpr{
									X:   translatedassertImportIdent,
									Sel: &ast.Ident{Name: "RVOf"},
								},
								Args: []ast.Expr{n.Args[0]},
							},
							Sel: &ast.Ident{Name: "Convert"},
						},
						Args: []ast.Expr{
							createReflectTypeExprFromTypeExpr(n.Fun),
						},
					},
				},
			}

			ps = append(ps, newPrintExpr(n.Pos()-offset, newExpr))
			ps = append(ps, argsPrints...)

			*n = *newExpr
		} else {
			argsPrints := []printExpr{}
			for _, arg := range n.Args {
				argsPrints = append(argsPrints, extractPrintExprs(filename, line, offset, n, arg)...)
			}

			var memorized *ast.CallExpr

			if p, ok := parent.(*ast.CallExpr); ok && isAssert(assertImportIdent, p) {
				memorized = createMemorizedFuncCall(filename, line, n, "Bool")
			} else {
				memorized = createMemorizedFuncCall(filename, line, n, "Interface")
			}

			ps = append(ps, newPrintExpr(n.Pos()-offset, memorized))
			ps = append(ps, argsPrints...)

			*n = *memorized
		}
	case *ast.SliceExpr:
		n := n.(*ast.SliceExpr)
		ps = append(ps, extractPrintExprs(filename, line, offset, n, n.Low)...)
		ps = append(ps, extractPrintExprs(filename, line, offset, n, n.High)...)
		if n.Slice3 {
			ps = append(ps, extractPrintExprs(filename, line, offset, n, n.Max)...)
		}
	}

	return ps
}

func replaceBinaryExpr(parent ast.Node, oldExpr *ast.BinaryExpr, newExpr ast.Expr) {
	switch parent.(type) {
	case *ast.CallExpr:
		parent := parent.(*ast.CallExpr)
		for i, arg := range parent.Args {
			if arg == oldExpr {
				parent.Args[i] = newExpr
				return
			}
		}
	case *ast.KeyValueExpr:
		parent := parent.(*ast.KeyValueExpr)
		switch oldExpr {
		case parent.Key:
			parent.Key = newExpr
			return
		case parent.Value:
			parent.Value = newExpr
			return
		}
	case *ast.IndexExpr:
		parent := parent.(*ast.IndexExpr)
		if parent.Index == oldExpr {
			parent.Index = newExpr
			return
		}
	case *ast.ParenExpr:
		parent := parent.(*ast.ParenExpr)
		if parent.X == oldExpr {
			parent.X = newExpr
			return
		}
	case *ast.BinaryExpr:
		parent := parent.(*ast.BinaryExpr)
		switch oldExpr {
		case parent.X:
			parent.X = newExpr
			return
		case parent.Y:
			parent.Y = newExpr
			return
		}
	}

	panic("[gnewExprwt]Unexpected Error on replacing *ast.BinaryExpr by translatedassert.Op*()")
}

func getTypeInfo(importDir string, fset *token.FileSet, files []*ast.File) *types.Info {
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
	typesConfig.Import = gcimporter.Import

	pkg := types.NewPackage(importDir, "")
	info := &types.Info{
		Types:      map[ast.Expr]types.TypeAndValue{},
		Defs:       map[*ast.Ident]types.Object{},
		Uses:       map[*ast.Ident]types.Object{},
		Implicits:  map[ast.Node]types.Object{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
		Scopes:     map[ast.Node]*types.Scope{},
		InitOrder:  []*types.Initializer{},
	}
	err := types.NewChecker(&typesConfig, fset, pkg, info).Files(files)
	if err != nil {
		fmt.Println("Error in checker: " + importDir)
		fmt.Println(err.Error())
		panic(err)
	}

	return info
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

func replaceAllRawStringLitByStringLit(root ast.Node) {
	ast.Inspect(root, func(n ast.Node) bool {
		if n, ok := n.(*ast.BasicLit); ok {
			if isRawStringLit(n) {
				n.Value = strconv.Quote(strings.Trim(n.Value, "`"))
			}
		}

		return true
	})
}

func getAssertImport(a *ast.File) *ast.ImportSpec {
	for _, decl := range a.Decls {
		decl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		if len(decl.Specs) == 0 {
			continue
		}

		if _, ok := decl.Specs[0].(*ast.ImportSpec); !ok {
			continue
		}

		for _, imp := range decl.Specs {
			imp := imp.(*ast.ImportSpec)

			if imp.Path.Value == `"github.com/ToQoz/gopwt/assert"` {
				return imp
			}
		}
	}

	return nil
}

// FIXME
func isGoFile2(name string) bool {
	return strings.HasSuffix(name, ".go") && !strings.HasPrefix(name, ".") && !strings.HasPrefix(name, "_")
}

func isTestGoFile(name string) bool {
	return strings.HasSuffix(name, "_test.go") && !strings.HasPrefix(name, ".") && !strings.HasPrefix(name, "_")
}

func isAssert(x *ast.Ident, c *ast.CallExpr) bool {
	if s, ok := c.Fun.(*ast.SelectorExpr); ok {
		return s.X.(*ast.Ident).Name == x.Name && s.Sel.Name == "OK"
	}

	return false
}

func isBuiltinFunc(n *ast.CallExpr) bool {
	if f, ok := n.Fun.(*ast.Ident); ok {
		for _, b := range builtinFuncs {
			if f.Name == b {
				return true
			}
		}
	}

	return false
}

func isMapType(n ast.Node) bool {
	if n, ok := n.(*ast.CompositeLit); ok {
		_, ismap := n.Type.(*ast.MapType)
		return ismap
	}

	return false
}

func isRawStringLit(n *ast.BasicLit) bool {
	return n.Kind == token.STRING && strings.HasPrefix(n.Value, "`") && strings.HasSuffix(n.Value, "`")
}

// --------------------------------------------------------------------------------
// AST generating shortcuts
// --------------------------------------------------------------------------------

func createUntypedCallExprFromBuiltinCallExpr(n *ast.CallExpr) *ast.CallExpr {
	createAltBuiltin := func(bfuncName string, args []ast.Expr) *ast.CallExpr {
		return &ast.CallExpr{
			Fun:  &ast.SelectorExpr{X: translatedassertImportIdent, Sel: &ast.Ident{Name: "B" + bfuncName}},
			Args: args,
		}
	}

	name := n.Fun.(*ast.Ident).Name

	switch name {
	case "append", "cap", "complex", "copy", "imag", "len", "real":
		return createAltBuiltin(name, n.Args)
	case "new":
		return createAltBuiltin(name, []ast.Expr{createReflectTypeExprFromTypeExpr(n.Args[0])})
	case "make":
		args := []ast.Expr{}
		args = append(args, createReflectTypeExprFromTypeExpr(n.Args[0]))
		args = append(args, n.Args[1:]...)
		return createAltBuiltin(name, args)
	default:
		panic(fmt.Errorf("%s can't be used in assert", name))
	}
}

func createReflectTypeExprFromTypeExpr(n ast.Expr) ast.Expr {
	canUseCompositeLit := true

	if n, ok := n.(*ast.Ident); ok {
		switch n.Name {
		case "string", "rune",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"int8", "int32", "int64", "int",
			"float32", "float64",
			"complex64", "complex128",
			"bool", "uintptr", "error":

			canUseCompositeLit = false
		}
	}

	if _, ok := n.(*ast.ChanType); ok {
		canUseCompositeLit = false
	}

	if !canUseCompositeLit {
		return &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X:   translatedassertImportIdent,
								Sel: &ast.Ident{Name: "RVOf"},
							},
							Args: []ast.Expr{
								&ast.CallExpr{
									Fun:  &ast.Ident{Name: "new"},
									Args: []ast.Expr{n},
								},
							},
						},
						Sel: &ast.Ident{Name: "Elem"},
					},
				},
				Sel: &ast.Ident{Name: "Type"},
			},
		}
	}

	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   translatedassertImportIdent,
			Sel: &ast.Ident{Name: "RTOf"},
		},
		Args: []ast.Expr{&ast.CompositeLit{Type: n}},
	}
}

// createUntypedExprFromBinaryExpr creates untyped operator-func(translatedassert.Op*()) from BinaryExpr
// if given BinaryExpr is untyped, returns it.
func createUntypedExprFromBinaryExpr(n *ast.BinaryExpr) ast.Expr {
	createFuncOp := func(opName string, x ast.Expr, y ast.Expr) *ast.CallExpr {
		return &ast.CallExpr{
			Fun:  &ast.SelectorExpr{X: translatedassertImportIdent, Sel: &ast.Ident{Name: "Op" + opName}},
			Args: []ast.Expr{x, y},
		}
	}

	// http://golang.org/ref/spec#Operators_and_Delimiters
	// +    sum                    integers, floats, complex values, strings
	// -    difference             integers, floats, complex values
	// *    product                integers, floats, complex values
	// /    quotient               integers, floats, complex values
	// %    remainder              integers

	// &    bitwise AND            integers
	// |    bitwise OR             integers
	// ^    bitwise XOR            integers
	// &^   bit clear (AND NOT)    integers

	// <<   left shift             integer << unsigned integer
	// >>   right shift            integer >> unsigned integer

	// http://golang.org/ref/spec#Logical_operators
	// Logical operators apply to boolean values and yield a result of the same type as the operands. The right operand is evaluated conditionally.

	// &&    conditional AND    p && q  is  "if p then q else false"
	// ||    conditional OR     p || q  is  "if p then true else q"
	switch n.Op {
	case token.ADD: // +
		return createFuncOp("ADD", n.X, n.Y)
	case token.SUB: // -
		return createFuncOp("SUB", n.X, n.Y)
	case token.MUL: // *
		return createFuncOp("MUL", n.X, n.Y)
	case token.QUO: // /
		return createFuncOp("QUO", n.X, n.Y)
	case token.REM: // %
		return createFuncOp("REM", n.X, n.Y)
	case token.AND: // &
		return createFuncOp("AND", n.X, n.Y)
	case token.OR: // |
		return createFuncOp("OR", n.X, n.Y)
	case token.XOR: // ^
		return createFuncOp("XOR", n.X, n.Y)
	case token.AND_NOT: // &^
		return createFuncOp("ANDNOT", n.X, n.Y)
	case token.SHL: // <<
		return createFuncOp("SHL", n.X, n.Y)
	case token.SHR: // >>
		return createFuncOp("SHR", n.X, n.Y)
	case token.LAND: // &&
		return createFuncOp("LAND", n.X, n.Y)
	case token.LOR: // ||
		return createFuncOp("LOR", n.X, n.Y)
	}

	return n
}

// f(a, b) -> translatedassert.FRVInterface(translatedassert.MFCall(filename, line, pos, f, translatedassert.RVOf(a), translatedassert.RVOf(b)))
func createMemorizedFuncCall(filename string, line int, n *ast.CallExpr, returnType string) *ast.CallExpr {
	c := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   translatedassertImportIdent,
			Sel: &ast.Ident{Name: "MFCall"},
		},
		Args: []ast.Expr{
			&ast.BasicLit{Kind: token.STRING, Value: strconv.Quote(filename)},
			&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(line)},
			&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(int(n.Pos()))},
			&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   translatedassertImportIdent,
					Sel: &ast.Ident{Name: "RVOf"},
				},
				Args: []ast.Expr{n.Fun},
			},
		},
	}

	args := []ast.Expr{}
	for _, a := range n.Args {
		args = append(args, &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   translatedassertImportIdent,
				Sel: &ast.Ident{Name: "RVOf"},
			},
			Args: []ast.Expr{a},
		})
	}
	c.Args = append(c.Args, args...)

	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   translatedassertImportIdent,
			Sel: &ast.Ident{Name: "FRV" + returnType},
		},
		Args: []ast.Expr{
			c,
		},
	}
}

func createPosValuePairExpr(ps []printExpr) []ast.Expr {
	args := []ast.Expr{}

	for _, n := range ps {
		a := &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   translatedassertImportIdent,
				Sel: &ast.Ident{Name: "NewPosValuePair"},
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(int(n.Pos))},
				n.Expr,
			},
		}

		args = append(args, a)
	}

	return args
}

func createRawStringLit(s string) *ast.BasicLit {
	return &ast.BasicLit{Kind: token.STRING, Value: "`" + s + "`"}
}
