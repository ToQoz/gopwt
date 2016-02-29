package main

import (
	"bytes"
	"go/ast"
	"go/build"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/types"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	translatedassertImportIdent = &ast.Ident{Name: "translatedassert"}
	assertImportIdent           = &ast.Ident{Name: "assert"}
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
	installable := false

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

		if !strings.HasSuffix(path, "_test.go") {
			installable = true
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

	deps, err := findDeps(importPath, tempGoSrcDir)
	if err != nil {
		return err
	}
	if installable {
		deps = append(deps, ".")
	}

	if len(deps) > 0 {
		// Create binary before type assuming from ast.Node
		install := exec.Command("go", "install")
		install.Dir = pkgDir
		install.Stdout = os.Stdout
		install.Stderr = os.Stderr
		if *verbose {
			install.Args = append(install.Args, "-v")
		}

		install.Args = append(install.Args, deps...)
		err = install.Run()
		if err != nil {
			return err
		}
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

func findDeps(importPath, srcDir string) ([]string, error) {
	deps := []string{}

	pkg, err := build.Import(importPath, srcDir, build.AllowBinary)
	if err != nil {
		return nil, err
	}

	for _, imp := range pkg.Imports {
		if imp == importPath {
			continue
		}
		deps = append(deps, imp)
	}

	for _, imp := range pkg.TestImports {
		if imp == importPath {
			continue
		}

		f := false
		for _, arg := range deps {
			if arg == imp {
				f = true
			}
		}

		if !f {
			deps = append(deps, imp)
		}
	}
	return deps, nil
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
			for _, tdata := range strings.Split(*testdata, ",") {
				if strings.Split(pathFromImportDir, string(filepath.Separator))[0] == tdata {
					di, err := os.Stat(filepath.Dir(path))
					if err != nil {
						return err
					}
					err = os.Mkdir(outPath, di.Mode())
					if err != nil {
						return err
					}
					return nil
				}
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

			rewriteAssert(fset.File(n.Pos()), n)
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

// rewriteAssert rewrites assert to translatedassert
func rewriteAssert(file *token.File, n *ast.CallExpr) {
	filename := file.Name()
	line := file.Line(n.Pos())

	b := []byte{}
	buf := bytes.NewBuffer(b)
	// printing valid expr must success
	must(printer.Fprint(buf, token.NewFileSet(), n))
	originalExprString := buf.String()

	// OK(t, a == b, "message", "message-2")
	// ---> OK(t, a == b, []string{"messages", "message-2"})
	testingT := n.Args[0]
	testExpr := n.Args[1]
	messages := createArrayTypeCompositLit("string")
	if len(n.Args) > 2 {
		for _, msg := range n.Args[2:] {
			messages.Elts = append(messages.Elts, msg)
		}
	}
	n.Args = []ast.Expr{testingT}
	n.Args = append(n.Args, testExpr)
	n.Args = append(n.Args, messages)

	// header
	n.Args = append(n.Args, createRawStringLit("FAIL"))
	// filename
	n.Args = append(n.Args, createRawStringLit(filename))
	// line
	n.Args = append(n.Args, &ast.BasicLit{Value: strconv.Itoa(line), Kind: token.INT})
	// string of original expr
	n.Args = append(n.Args, createRawStringLit(originalExprString))
	// terminal width
	n.Args = append(n.Args, &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(termw)})
	// pos-value pairs
	n.Args = append(n.Args, createPosValuePairExpr(extractPrintExprs(filename, line, n.Pos()-1, n, n.Args[1]))...)
	n.Fun.(*ast.SelectorExpr).X = &ast.Ident{Name: "translatedassert"}
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
		// HACK:
		// skip ident type is invalid
		// e.g. pkg.ErrNoRows
		//      ^^^
		// I'm searchng more better ways....
		if typesInfo != nil {
			if basic, ok := typesInfo.TypeOf(n).(*types.Basic); ok && basic.Kind() == types.Invalid {
				return ps
			}
		}
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

		n.X = createReflectBoolExpr(createReflectValueOfExpr(n.X))

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

			newExpr := createReflectInterfaceExpr(&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   createReflectValueOfExpr(n.Args[0]),
					Sel: &ast.Ident{Name: "Convert"},
				},
				Args: []ast.Expr{
					createReflectTypeExprFromTypeExpr(n.Fun),
				},
			})

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
