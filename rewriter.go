package main

import (
	"bytes"
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
)

func rewritePackage(dirPath, importPath string, tempGoSrcDir string) error {
	err := filepath.Walk(dirPath, func(path string, fInfo os.FileInfo, err error) error {
		if fInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil
		}

		pathFromImportDir, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		if fInfo.IsDir() {
			rel, err := filepath.Rel(dirPath, path)
			if err != nil {
				return err
			}

			if rel == "." {
				return nil
			}

			// copy all files in <dirPath>/testdata/**/*
			if strings.Split(rel, "/")[0] == "testdata" {
				err = os.MkdirAll(filepath.Join(tempGoSrcDir, importPath, pathFromImportDir), os.ModePerm)
				if err != nil {
					return err
				}

				return nil
			}

			return filepath.SkipDir
		}

		out, err := os.Create(filepath.Join(tempGoSrcDir, importPath, pathFromImportDir))
		if err != nil {
			return err
		}
		defer func() {
			out.Close()
		}()

		err = rewriteFile(path, importPath, out)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func rewriteFile(path, importPath string, out io.Writer) error {
	translatedassertImportIdent = &ast.Ident{Name: "translatedassert"}
	assertImportIdent = &ast.Ident{Name: "assert"}

	copyFile := func() error {
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

	if !isTestGoFile(path) {
		return copyFile()
	}

	fset := token.NewFileSet()
	a, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return err
	}

	err = translateAssertImport(a)
	if err != nil {
		if err == errAssertImportNotFound {
			return copyFile()
		}

		return err
	}

	err = translateAllAsserts(fset, a)
	if err != nil {
		fmt.Println("error")
		return err
	}

	err = printer.Fprint(out, fset, a)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func isTestGoFile(name string) bool {
	return strings.HasSuffix(name, "_test.go") && !strings.HasPrefix(name, ".") && !strings.HasPrefix(name, "_")
}

func newPackageInfo(globalOrLocalImportPath string) (*packageInfo, error) {
	var err error
	var importPath string
	var dirPath string
	var recursive bool

	if strings.HasSuffix(globalOrLocalImportPath, "/...") {
		recursive = true
		globalOrLocalImportPath = strings.TrimSuffix(globalOrLocalImportPath, "/...")
	}

	if globalOrLocalImportPath == "" {
		globalOrLocalImportPath = "."
	}

	if strings.HasPrefix(globalOrLocalImportPath, ".") {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		dirPath = filepath.Join(wd, globalOrLocalImportPath)
		if _, err := os.Stat(dirPath); err != nil {
			return nil, err
		}

		importPath, err = findImportPathByPath(dirPath)
		if err != nil {
			return nil, err
		}
	} else {
		importPath = globalOrLocalImportPath

		dirPath, err = findPathByImportPath(importPath)
		if err != nil {
			return nil, err
		}
	}

	return &packageInfo{dirPath: dirPath, importPath: importPath, recursive: recursive}, nil
}

func translateAssertImport(a *ast.File) error {
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

			if imp.Path.Value != `"github.com/ToQoz/gopwt/assert"` {
				continue
			}

			imp.Path.Value = `"github.com/ToQoz/gopwt/translatedassert"`
			if imp.Name != nil {
				assertImportIdent = imp.Name
				imp.Name = translatedassertImportIdent
			}

			goto Done
		}
	}

	return errAssertImportNotFound

Done:
	return nil
}

func translateAllAsserts(fset *token.FileSet, a ast.Node) error {
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
			// header := fmt.Sprintf("[FAIL] %s:%d", filename, line)

			replaceAllRawStringLitByStringLit(n)

			b := []byte{}
			buf := bytes.NewBuffer(b)
			// This printing **must** success.(valid expr -> code)
			// So call panic on failing.
			err := printer.Fprint(buf, token.NewFileSet(), n)
			if err != nil {
				fmt.Println(err.Error())
				panic(err)
			}

			// This parsing **must** success.(valid code -> expr)
			// So call panic on failing.
			formatted, err := parser.ParseExpr(buf.String())
			if err != nil {
				fmt.Println(err.Error())
				panic(err)
			}
			*n = *formatted.(*ast.CallExpr)

			n.Args = append(n.Args, createRawStringLit("FAIL"))
			n.Args = append(n.Args, createRawStringLit(filename))
			n.Args = append(n.Args, &ast.BasicLit{Value: strconv.Itoa(line), Kind: token.INT})
			n.Args = append(n.Args, createRawStringLit(buf.String()))
			n.Args = append(n.Args, &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(termw)})
			n.Args = append(n.Args, createPosValuePairExpr(extractPrintExprs(filename, line, n, n.Args[1]))...)
			n.Fun.(*ast.SelectorExpr).X = &ast.Ident{Name: "translatedassert"}
		}

		return true
	})

	return nil
}

type printExpr struct {
	Pos  int
	Expr ast.Expr
}

func newPrintExpr(pos token.Pos, e ast.Expr) printExpr {
	return printExpr{Pos: int(pos), Expr: e}
}

func extractPrintExprs(filename string, line int, parent ast.Expr, n ast.Expr) []printExpr {
	ps := []printExpr{}

	switch n.(type) {
	case *ast.BasicLit:
		n := n.(*ast.BasicLit)
		if n.Kind == token.STRING {
			if len(strings.Split(n.Value, "\\n")) > 1 {
				ps = append(ps, newPrintExpr(n.Pos(), n))
			}
		}
	case *ast.CompositeLit:
		n := n.(*ast.CompositeLit)

		for _, elt := range n.Elts {
			ps = append(ps, extractPrintExprs(filename, line, n, elt)...)
		}
	case *ast.KeyValueExpr:
		n := n.(*ast.KeyValueExpr)

		if isMapType(parent) {
			ps = append(ps, extractPrintExprs(filename, line, n, n.Key)...)
		}

		ps = append(ps, extractPrintExprs(filename, line, n, n.Value)...)
	case *ast.Ident:
		ps = append(ps, newPrintExpr(n.Pos(), n))
	case *ast.ParenExpr:
		n := n.(*ast.ParenExpr)
		ps = append(ps, extractPrintExprs(filename, line, n, n.X)...)
	case *ast.StarExpr:
		n := n.(*ast.StarExpr)
		ps = append(ps, newPrintExpr(n.Pos(), n))
		ps = append(ps, extractPrintExprs(filename, line, n, n.X)...)
	case *ast.UnaryExpr:
		n := n.(*ast.UnaryExpr)
		x := extractPrintExprs(filename, line, n, n.X)

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

		ps = append(ps, newPrintExpr(n.Pos(), n))
		ps = append(ps, x...)
	case *ast.BinaryExpr:
		n := n.(*ast.BinaryExpr)

		var x, y []printExpr

		x = extractPrintExprs(filename, line, n, n.X)
		y = extractPrintExprs(filename, line, n, n.Y)

		newExpr := replaceBinaryExprInParent(parent, n)
		ps = append(ps, x...)
		ps = append(ps, newPrintExpr(n.OpPos, newExpr))
		ps = append(ps, y...)
	case *ast.IndexExpr:
		n := n.(*ast.IndexExpr)
		ps = append(ps, extractPrintExprs(filename, line, n, n.X)...)
		ps = append(ps, extractPrintExprs(filename, line, n, n.Index)...)
	case *ast.SelectorExpr:
		n := n.(*ast.SelectorExpr)
		ps = append(ps, extractPrintExprs(filename, line, n, n.X)...)
		ps = append(ps, newPrintExpr(n.Sel.Pos(), n))
	case *ast.CallExpr:
		n := n.(*ast.CallExpr)
		if isBuiltinFunc(n) { // don't memorize buildin methods
			ps = append(ps, newPrintExpr(n.Pos(), n))
			for _, arg := range n.Args {
				ps = append(ps, extractPrintExprs(filename, line, n, arg)...)
			}
		} else {
			var memorized *ast.CallExpr

			if p, ok := parent.(*ast.CallExpr); ok && isAssert(assertImportIdent, p) {
				memorized = createMemorizedFuncCall(filename, line, n, "Bool")
			} else {
				memorized = createMemorizedFuncCall(filename, line, n, "Interface")
			}

			ps = append(ps, newPrintExpr(n.Pos(), memorized))
			for _, arg := range n.Args {
				ps = append(ps, extractPrintExprs(filename, line, nil, arg)...)
			}

			*n = *memorized
		}
	case *ast.SliceExpr:
		n := n.(*ast.SliceExpr)
		ps = append(ps, extractPrintExprs(filename, line, n, n.Low)...)
		ps = append(ps, extractPrintExprs(filename, line, n, n.High)...)
		if n.Slice3 {
			ps = append(ps, extractPrintExprs(filename, line, n, n.Max)...)
		}
	}

	return ps
}

// replaceBinaryExprInParent replaces binaryExpr by operator-func(impled by translatedassert)'s callExpr if its operator can't take interface{}
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
func replaceBinaryExprInParent(parent ast.Node, n *ast.BinaryExpr) ast.Expr {
	replace := func(parent ast.Node, n *ast.BinaryExpr, newExpr ast.Expr) {
		switch parent.(type) {
		case *ast.CallExpr:
			parent := parent.(*ast.CallExpr)
			for i, arg := range parent.Args {
				if n == arg {
					parent.Args[i] = newExpr
					return
				}
			}
		case *ast.KeyValueExpr:
			parent := parent.(*ast.KeyValueExpr)
			if parent.Key == n {
				parent.Key = newExpr
				return
			} else if parent.Value == n {
				parent.Value = newExpr
				return
			}
		case *ast.IndexExpr:
			parent := parent.(*ast.IndexExpr)
			if parent.Index == n {
				parent.Index = newExpr
				return
			}
		case *ast.ParenExpr:
			parent := parent.(*ast.ParenExpr)
			if parent.X == n {
				parent.X = newExpr
				return
			}
		case *ast.BinaryExpr:
			parent := parent.(*ast.BinaryExpr)
			if parent.X == n {
				parent.X = newExpr
				return
			} else if parent.Y == n {
				parent.Y = newExpr
				return
			}
		default:
			panic("[gnewExprwt]Unexpected Error on replacing *ast.BinaryExpr by translatedassert.Op*()")
		}
	}

	createFuncOp := func(opName string, x ast.Expr, y ast.Expr) *ast.CallExpr {
		return &ast.CallExpr{
			Fun:  &ast.SelectorExpr{X: translatedassertImportIdent, Sel: &ast.Ident{Name: "Op" + opName}},
			Args: []ast.Expr{x, y},
		}
	}

	var newExpr ast.Expr

	switch n.Op {
	case token.ADD: // +
		newExpr = createFuncOp("ADD", n.X, n.Y)
		replace(parent, n, newExpr)
	case token.SUB: // -
		newExpr = createFuncOp("SUB", n.X, n.Y)
		replace(parent, n, newExpr)
	case token.MUL: // *
		newExpr = createFuncOp("MUL", n.X, n.Y)
		replace(parent, n, newExpr)
	case token.QUO: // /
		newExpr = createFuncOp("QUO", n.X, n.Y)
		replace(parent, n, newExpr)
	case token.REM: // %
		newExpr = createFuncOp("REM", n.X, n.Y)
		replace(parent, n, newExpr)
	case token.AND: // &
		newExpr = createFuncOp("AND", n.X, n.Y)
		replace(parent, n, newExpr)
	case token.OR: // |
		newExpr = createFuncOp("OR", n.X, n.Y)
		replace(parent, n, newExpr)
	case token.XOR: // ^
		newExpr = createFuncOp("XOR", n.X, n.Y)
		replace(parent, n, newExpr)
	case token.AND_NOT: // &^
		newExpr = createFuncOp("ANDNOT", n.X, n.Y)
		replace(parent, n, newExpr)
	case token.SHL: // <<
		newExpr = createFuncOp("SHL", n.X, n.Y)
		replace(parent, n, newExpr)
	case token.SHR: // >>
		newExpr = createFuncOp("SHR", n.X, n.Y)
		replace(parent, n, newExpr)
	case token.LAND: // &&
		newExpr = createFuncOp("LAND", n.X, n.Y)
		replace(parent, n, newExpr)
	case token.LOR: // ||
		newExpr = createFuncOp("LOR", n.X, n.Y)
		replace(parent, n, newExpr)
	default:
		newExpr = n
	}

	return newExpr
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
	return strings.HasPrefix(n.Value, "`") && strings.HasSuffix(n.Value, "`")
}

// --------------------------------------------------------------------------------
// AST generating shortcuts
// --------------------------------------------------------------------------------

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
