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
		return err
	}

	err = printer.Fprint(out, fset, a)
	if err != nil {
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

func translateAllAsserts(fset *token.FileSet, a *ast.File) error {
	ast.Inspect(a, func(n ast.Node) bool {
		if n, ok := n.(*ast.CallExpr); ok {
			// skip inspecting children in n
			if !isAssert(assertImportIdent, n) {
				return false
			}

			file := fset.File(n.Pos())
			header := fmt.Sprintf("[FAIL] %s:%d", file.Name(), file.Line(n.Pos()))

			replaceAllRawStringLitByStringLit(n)

			b := []byte{}
			buf := bytes.NewBuffer(b)
			// This printing **must** success.(valid expr -> code)
			// So call panic on failing.
			err := printer.Fprint(buf, token.NewFileSet(), n)
			if err != nil {
				panic(err)
			}

			// This parsing **must** success.(valid code -> expr)
			// So call panic on failing.
			formatted, err := parser.ParseExpr(buf.String())
			if err != nil {
				panic(err)
			}
			*n = *formatted.(*ast.CallExpr)

			n.Args = append(n.Args, createRawStringLit(header+"\n"+buf.String()))
			n.Args = append(n.Args, &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(termw)})
			n.Args = append(n.Args, createPosValuePairExpr(extractPrintExprs(nil, n.Args[1]))...)
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

func extractPrintExprs(parent ast.Expr, n ast.Expr) []printExpr {
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
			ps = append(ps, extractPrintExprs(n, elt)...)
		}
	case *ast.KeyValueExpr:
		n := n.(*ast.KeyValueExpr)

		if isMapType(parent) {
			ps = append(ps, extractPrintExprs(n, n.Key)...)
		}

		ps = append(ps, extractPrintExprs(n, n.Value)...)
	case *ast.Ident:
		ps = append(ps, newPrintExpr(n.Pos(), n))
	case *ast.ParenExpr:
		n := n.(*ast.ParenExpr)
		ps = append(ps, extractPrintExprs(n, n.X)...)
	case *ast.StarExpr:
		n := n.(*ast.StarExpr)
		ps = append(ps, newPrintExpr(n.Pos(), n))
		ps = append(ps, extractPrintExprs(n, n.X)...)
	case *ast.UnaryExpr:
		n := n.(*ast.UnaryExpr)
		ps = append(ps, newPrintExpr(n.Pos(), n))
		ps = append(ps, extractPrintExprs(n, n.X)...)
	case *ast.BinaryExpr:
		n := n.(*ast.BinaryExpr)
		ps = append(ps, extractPrintExprs(n, n.X)...)
		ps = append(ps, newPrintExpr(n.OpPos, n))
		ps = append(ps, extractPrintExprs(n, n.Y)...)
	case *ast.IndexExpr:
		n := n.(*ast.IndexExpr)
		ps = append(ps, extractPrintExprs(n, n.X)...)
		ps = append(ps, extractPrintExprs(n, n.Index)...)
	case *ast.SelectorExpr:
		n := n.(*ast.SelectorExpr)
		ps = append(ps, extractPrintExprs(n, n.X)...)
		ps = append(ps, newPrintExpr(n.Sel.Pos(), n))
	case *ast.CallExpr:
		n := n.(*ast.CallExpr)

		ps = append(ps, newPrintExpr(n.Pos(), n))
		for _, arg := range n.Args {
			ps = append(ps, extractPrintExprs(n, arg)...)
		}
	}

	return ps
}

func isAssert(x *ast.Ident, c *ast.CallExpr) bool {
	if s, ok := c.Fun.(*ast.SelectorExpr); ok {
		return s.X.(*ast.Ident).Name == x.Name && s.Sel.Name == "OK"
	}

	return false
}

func createRawStringLit(s string) *ast.BasicLit {
	return &ast.BasicLit{Kind: token.STRING, Value: "`" + s + "`"}
}

func isRawStringLit(n *ast.BasicLit) bool {
	return strings.HasPrefix(n.Value, "`") && strings.HasSuffix(n.Value, "`")
}

func isMapType(n ast.Node) bool {
	if n, ok := n.(*ast.CompositeLit); ok {
		_, ismt := n.Type.(*ast.MapType)
		return ismt
	}

	return false
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
