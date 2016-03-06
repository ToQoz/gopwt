package main

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/go/types"
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
	originalFset := token.NewFileSet()
	origFiles := []*ast.File{}
	root := filepath.Join(tempGoSrcDir, importPath)

	err = filepath.Walk(root, func(path string, fInfo os.FileInfo, err error) error {
		if fInfo.IsDir() {
			if path == root {
				return nil
			}

			return filepath.SkipDir
		}

		if !isGoFileName(path) {
			return nil
		}

		// parse copyed & normalized file. e.g. multi-line CompositeLit -> single-line
		a, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		files = append(files, a)

		// parse original file
		o, err := parser.ParseFile(originalFset, filepath.Join(pkgDir, filepath.Base(path)), nil, 0)
		if err != nil {
			return err
		}
		origFiles = append(origFiles, o)

		return nil
	})
	if err != nil {
		return err
	}

	typesInfo, err := getTypeInfo(pkgDir, importPath, tempGoSrcDir, fset, files)
	if err != nil {
		return err
	}

	// Rewrite files
	for i, f := range files {
		path := fset.File(f.Package).Name()

		if !isTestGoFileName(path) {
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
		err = rewriteFile(typesInfo, fset, originalFset, f, origFiles[i], out)
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

		outOrig, err := os.OpenFile(outPath+".orig", os.O_RDWR|os.O_CREATE, fInfo.Mode())
		if err != nil {
			return err
		}
		defer outOrig.Close()

		if !isTestGoFileName(path) {
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
		inspectAssert(a, func(n *ast.CallExpr) {
			// 1. replace rawStringLit to stringLit
			replaceAllRawStringLitByStringLit(n)
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

func rewriteFile(typesInfo *types.Info, fset, originalFset *token.FileSet, file, origFile *ast.File, out io.Writer) error {
	assertPositions := []token.Position{}
	inspectAssert(origFile, func(n *ast.CallExpr) {
		assertPositions = append(assertPositions, originalFset.Position(n.Pos()))
	})
	i := 0
	inspectAssert(file, func(n *ast.CallExpr) {
		pos := assertPositions[i]
		i++
		rewriteAssert(typesInfo, pos, n)
	})

	err := printer.Fprint(out, fset, file)
	if err != nil {
		return err
	}
	return nil
}

// rewriteAssert rewrites assert to translatedassert
func rewriteAssert(typesInfo *types.Info, position token.Position, n *ast.CallExpr) {
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

	posOffset := n.Pos() - 1
	// expected pos-value index
	// got pos-value index
	expectedPosValueIndex := -1
	gotPosValueIndex := -1
	// unknown
	if isEqualExpr(testExpr) {
		bin := testExpr.(*ast.BinaryExpr)
		expectedPosValueIndex = int(resultPosOf(bin.Y) - posOffset)
		gotPosValueIndex = int(resultPosOf(bin.X) - posOffset)
	} else if isReflectDeepEqual(testExpr) {
		call := testExpr.(*ast.CallExpr)
		expectedPosValueIndex = int(resultPosOf(call.Args[1]) - posOffset)
		gotPosValueIndex = int(resultPosOf(call.Args[0]) - posOffset)
	}

	// header
	n.Args = append(n.Args, createRawStringLit("FAIL"))
	// filename
	n.Args = append(n.Args, createRawStringLit(position.Filename))
	// line
	n.Args = append(n.Args, &ast.BasicLit{Value: strconv.Itoa(position.Line), Kind: token.INT})
	// string of original expr
	n.Args = append(n.Args, createRawStringLit(originalExprString))
	// terminal width
	n.Args = append(n.Args, &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(termw)})

	n.Args = append(
		n.Args,
		&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(expectedPosValueIndex)},
		&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(gotPosValueIndex)},
	)

	// pos-value pairs
	extractedPrintExprs := extractPrintExprs(typesInfo, position.Filename, position.Line, posOffset, n, n.Args[1])
	n.Args = append(n.Args, createPosValuePairExpr(extractedPrintExprs)...)
	n.Fun.(*ast.SelectorExpr).X = &ast.Ident{Name: "translatedassert"}
}

type printExpr struct {
	Pos          int
	Expr         ast.Expr
	OriginalExpr string
}

func newPrintExpr(pos token.Pos, newExpr ast.Expr, originalExpr string) printExpr {
	return printExpr{Pos: int(pos), Expr: newExpr, OriginalExpr: originalExpr}
}

func extractPrintExprs(typesInfo *types.Info, filename string, line int, offset token.Pos, parent ast.Expr, n ast.Expr) []printExpr {
	ps := []printExpr{}

	original := sprintCode(n)

	switch n.(type) {
	case *ast.BasicLit:
		n := n.(*ast.BasicLit)
		ps = append(ps, newPrintExpr(n.Pos()-offset, n, original))
	case *ast.CompositeLit:
		n := n.(*ast.CompositeLit)

		ps = append(ps, newPrintExpr(n.Pos()-offset, n, original))
		for _, elt := range n.Elts {
			ps = append(ps, extractPrintExprs(typesInfo, filename, line, offset, n, elt)...)
		}
	case *ast.KeyValueExpr:
		n := n.(*ast.KeyValueExpr)

		if isMapType(parent) {
			ps = append(ps, extractPrintExprs(typesInfo, filename, line, offset, n, n.Key)...)
		}

		ps = append(ps, extractPrintExprs(typesInfo, filename, line, offset, n, n.Value)...)
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
		ps = append(ps, newPrintExpr(n.Pos()-offset, n, original))
	case *ast.ParenExpr:
		n := n.(*ast.ParenExpr)
		ps = append(ps, extractPrintExprs(typesInfo, filename, line, offset, n, n.X)...)
	case *ast.StarExpr:
		n := n.(*ast.StarExpr)
		ps = append(ps, newPrintExpr(n.Pos()-offset, n, original))
		ps = append(ps, extractPrintExprs(typesInfo, filename, line, offset, n, n.X)...)
	case *ast.UnaryExpr:
		n := n.(*ast.UnaryExpr)
		x := extractPrintExprs(typesInfo, filename, line, offset, n, n.X)

		n.X = createReflectBoolExpr(createReflectValueOfExpr(n.X))

		ps = append(ps, newPrintExpr(n.Pos()-offset, n, original))
		ps = append(ps, x...)
	case *ast.BinaryExpr:
		n := n.(*ast.BinaryExpr)

		var x, y []printExpr

		x = extractPrintExprs(typesInfo, filename, line, offset, n, n.X)
		y = extractPrintExprs(typesInfo, filename, line, offset, n, n.Y)

		newExpr := createUntypedExprFromBinaryExpr(n)
		if newExpr != n {
			replaceBinaryExpr(parent, n, newExpr)
		}

		ps = append(ps, x...)
		ps = append(ps, newPrintExpr(n.OpPos-offset, newExpr, original))
		ps = append(ps, y...)
	case *ast.IndexExpr:
		n := n.(*ast.IndexExpr)
		// show value under the `[` in the same way as Groovy@v2.4.6 and power-assert-js@v1.2.0
		//   xs[i]
		//   | ||
		//   | |1
		//   | "b"
		//   ["a", "b"]
		ps = append(ps, extractPrintExprs(typesInfo, filename, line, offset, n, n.X)...)
		ps = append(ps, newPrintExpr(n.Index.Pos()-1-offset, n, original))
		ps = append(ps, extractPrintExprs(typesInfo, filename, line, offset, n, n.Index)...)
	case *ast.SelectorExpr:
		n := n.(*ast.SelectorExpr)
		ps = append(ps, extractPrintExprs(typesInfo, filename, line, offset, n, n.X)...)
		ps = append(ps, newPrintExpr(n.Sel.Pos()-offset, n, original))
	case *ast.CallExpr:
		n := n.(*ast.CallExpr)
		if isBuiltinFunc(n) { // don't memorize buildin methods
			newExpr := createUntypedCallExprFromBuiltinCallExpr(n)

			ps = append(ps, newPrintExpr(n.Pos()-offset, newExpr, original))
			for _, arg := range n.Args {
				ps = append(ps, extractPrintExprs(typesInfo, filename, line, offset, n, arg)...)
			}

			*n = *newExpr
		} else if isTypeConversion(typesInfo, n) {
			// T(v) can take only one argument.
			if len(n.Args) > 1 {
				panic("too many arguments for type conversion")
			} else if len(n.Args) < 1 {
				panic("missing argument for type conversion")
			}

			argsPrints := extractPrintExprs(typesInfo, filename, line, offset, n, n.Args[0])

			newExpr := createReflectInterfaceExpr(&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   createReflectValueOfExpr(n.Args[0]),
					Sel: &ast.Ident{Name: "Convert"},
				},
				Args: []ast.Expr{
					createReflectTypeExprFromTypeExpr(n.Fun),
				},
			})

			ps = append(ps, newPrintExpr(n.Pos()-offset, newExpr, original))
			ps = append(ps, argsPrints...)

			*n = *newExpr
		} else {
			argsPrints := []printExpr{}
			for _, arg := range n.Args {
				argsPrints = append(argsPrints, extractPrintExprs(typesInfo, filename, line, offset, n, arg)...)
			}

			var memorized *ast.CallExpr

			if p, ok := parent.(*ast.CallExpr); ok && isAssert(assertImportIdent, p) {
				memorized = createMemorizedFuncCall(filename, line, n, "Bool")
			} else {
				memorized = createMemorizedFuncCall(filename, line, n, "Interface")
			}

			ps = append(ps, newPrintExpr(n.Pos()-offset, memorized, original))
			ps = append(ps, argsPrints...)

			*n = *memorized
		}
	case *ast.SliceExpr:
		n := n.(*ast.SliceExpr)
		ps = append(ps, extractPrintExprs(typesInfo, filename, line, offset, n, n.Low)...)
		ps = append(ps, extractPrintExprs(typesInfo, filename, line, offset, n, n.High)...)
		if n.Slice3 {
			ps = append(ps, extractPrintExprs(typesInfo, filename, line, offset, n, n.Max)...)
		}
	}

	return ps
}

// "a" -> 0
// "obj.Fn" -> 3
// "1 + 2" -> 2
// "(1)" -> 1
func resultPosOf(n ast.Expr) token.Pos {
	switch n.(type) {
	case *ast.IndexExpr:
		n := n.(*ast.IndexExpr)
		return n.Index.Pos() - 1
	case *ast.SelectorExpr:
		n := n.(*ast.SelectorExpr)
		return resultPosOf(n.Sel)
	case *ast.BinaryExpr:
		n := n.(*ast.BinaryExpr)
		return n.OpPos
	case *ast.ParenExpr:
		n := n.(*ast.ParenExpr)
		return resultPosOf(n.X)
	default:
		return n.Pos()
	}
}
