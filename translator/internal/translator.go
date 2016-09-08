package internal

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	translatedAssertImportIdent = &ast.Ident{Name: "translatedassert"}
	AssertImportIdent           = &ast.Ident{Name: "assert"}
	Testdata                    = "testdata"
	TermWidth                   = 0
	Verbose                     = false
)

func Rewrite(gopath string, importpath, _filepath string, recursive bool) error {
	srcDir := filepath.Join(gopath, "src")

	err := filepath.Walk(_filepath, func(path string, fInfo os.FileInfo, err error) error {
		if fInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil
		}

		if !fInfo.IsDir() {
			return nil
		}

		files, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}
		if !ContainsGoFile(files) {
			// sub-packages maybe have gofiles, even if itself don't has gofiles
			if ContainsDirectory(files) {
				return nil
			}
			return filepath.SkipDir
		}

		rel, err := filepath.Rel(_filepath, path)
		if err != nil {
			return err
		}

		for _, tdata := range strings.Split(Testdata, ",") {
			if strings.Split(rel, "/")[0] == tdata {
				return filepath.SkipDir
			}
		}

		if rel != "." {
			if filepath.HasPrefix(rel, ".") {
				return filepath.SkipDir
			}

			if !recursive {
				return filepath.SkipDir
			}
		}

		importpath := filepath.Join(importpath, rel)

		err = os.MkdirAll(filepath.Join(srcDir, importpath), os.ModePerm)
		if err != nil {
			return err
		}

		err = rewritePackage(path, importpath, srcDir)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

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

		if !IsGoFileName(path) {
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

	typesInfo, err := GetTypeInfo(pkgDir, importPath, tempGoSrcDir, fset, files)
	if err != nil {
		return err
	}

	// Rewrite files
	for i, f := range files {
		path := fset.File(f.Package).Name()

		if !IsTestGoFileName(path) {
			continue
		}

		gopwtMainDropped := DropGopwtEmpower(f)

		AssertImportIdent = &ast.Ident{Name: "assert"}
		assertImport := GetAssertImport(f)
		if assertImport == nil {
			if !gopwtMainDropped {
				continue
			}
		} else {
			assertImport.Path.Value = `"github.com/ToQoz/gopwt/translatedassert"`

			if assertImport.Name != nil {
				AssertImportIdent = assertImport.Name
				assertImport.Name = translatedAssertImportIdent
			}
		}

		fi, err := os.Stat(path)
		if err != nil {
			return err
		}

		out, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, fi.Mode())
		defer out.Close()
		err = RewriteFile(typesInfo, fset, originalFset, f, origFiles[i], out)
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
			for _, tdata := range strings.Split(Testdata, ",") {
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

		if !IsTestGoFileName(path) {
			return CopyFile(path, out)
		}

		a, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}

		AssertImportIdent = &ast.Ident{Name: "assert"}
		assertImport := GetAssertImport(a)
		if assertImport == nil {
			return CopyFile(path, out)
		}
		if assertImport.Name != nil {
			AssertImportIdent = assertImport.Name
		}

		// Format and copy file
		//   - 1. replace rawStringLit to stringLit
		//   - 2. replace multiline compositLit to singleline one.
		//   - 3. copy
		InspectAssert(a, func(n *ast.CallExpr) {
			// 1. replace rawStringLit to stringLit
			ReplaceAllRawStringLitByStringLit(n)
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

func CopyFile(path string, out io.Writer) error {
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

func RewriteFile(typesInfo *types.Info, fset, originalFset *token.FileSet, file, origFile *ast.File, out io.Writer) error {
	assertPositions := []token.Position{}
	InspectAssert(origFile, func(n *ast.CallExpr) {
		assertPositions = append(assertPositions, originalFset.Position(n.Pos()))
	})
	i := 0
	InspectAssert(file, func(n *ast.CallExpr) {
		pos := assertPositions[i]
		i++
		RewriteAssert(typesInfo, pos, n)
	})

	err := printer.Fprint(out, fset, file)
	if err != nil {
		return err
	}
	return nil
}

// RewriteAssert rewrites assert to translatedassert
func RewriteAssert(typesInfo *types.Info, position token.Position, n *ast.CallExpr) {
	b := []byte{}
	buf := bytes.NewBuffer(b)
	// printing valid expr Must success
	Must(printer.Fprint(buf, token.NewFileSet(), n))
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
	if IsEqualExpr(testExpr) {
		bin := testExpr.(*ast.BinaryExpr)
		expectedPosValueIndex = int(ResultPosOf(bin.Y) - posOffset)
		gotPosValueIndex = int(ResultPosOf(bin.X) - posOffset)
	} else if IsReflectDeepEqual(testExpr) {
		call := testExpr.(*ast.CallExpr)
		expectedPosValueIndex = int(ResultPosOf(call.Args[1]) - posOffset)
		gotPosValueIndex = int(ResultPosOf(call.Args[0]) - posOffset)
	}

	// header
	n.Args = append(n.Args, CreateRawStringLit("FAIL"))
	// filename
	n.Args = append(n.Args, CreateRawStringLit(position.Filename))
	// line
	n.Args = append(n.Args, &ast.BasicLit{Value: strconv.Itoa(position.Line), Kind: token.INT})
	// string of original expr
	n.Args = append(n.Args, CreateRawStringLit(originalExprString))
	// terminal width
	n.Args = append(n.Args, &ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(TermWidth)})

	n.Args = append(
		n.Args,
		&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(expectedPosValueIndex)},
		&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(gotPosValueIndex)},
	)

	// pos-value pairs
	extractedPrintExprs := ExtractPrintExprs(typesInfo, position.Filename, position.Line, posOffset, n, n.Args[1])
	n.Args = append(n.Args, CreatePosValuePairExpr(extractedPrintExprs)...)
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

func ExtractPrintExprs(typesInfo *types.Info, filename string, line int, offset token.Pos, parent ast.Expr, n ast.Expr) []printExpr {
	ps := []printExpr{}

	original := SprintCode(n)

	switch n.(type) {
	case *ast.BasicLit:
		n := n.(*ast.BasicLit)
		ps = append(ps, newPrintExpr(n.Pos()-offset, n, original))
	case *ast.CompositeLit:
		n := n.(*ast.CompositeLit)

		ps = append(ps, newPrintExpr(n.Pos()-offset, n, original))
		for _, elt := range n.Elts {
			ps = append(ps, ExtractPrintExprs(typesInfo, filename, line, offset, n, elt)...)
		}
	case *ast.KeyValueExpr:
		n := n.(*ast.KeyValueExpr)

		if IsMapType(parent) {
			ps = append(ps, ExtractPrintExprs(typesInfo, filename, line, offset, n, n.Key)...)
		}

		ps = append(ps, ExtractPrintExprs(typesInfo, filename, line, offset, n, n.Value)...)
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
		ps = append(ps, ExtractPrintExprs(typesInfo, filename, line, offset, n, n.X)...)
	case *ast.StarExpr:
		n := n.(*ast.StarExpr)
		ps = append(ps, newPrintExpr(n.Pos()-offset, n, original))
		ps = append(ps, ExtractPrintExprs(typesInfo, filename, line, offset, n, n.X)...)
	case *ast.UnaryExpr:
		n := n.(*ast.UnaryExpr)
		x := ExtractPrintExprs(typesInfo, filename, line, offset, n, n.X)

		n.X = CreateReflectBoolExpr(CreateReflectValueOfExpr(n.X))

		ps = append(ps, newPrintExpr(n.Pos()-offset, n, original))
		ps = append(ps, x...)
	case *ast.BinaryExpr:
		n := n.(*ast.BinaryExpr)

		var x, y []printExpr

		x = ExtractPrintExprs(typesInfo, filename, line, offset, n, n.X)
		y = ExtractPrintExprs(typesInfo, filename, line, offset, n, n.Y)

		newExpr := CreateUntypedExprFromBinaryExpr(n)
		if newExpr != n {
			ReplaceBinaryExpr(parent, n, newExpr)
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
		ps = append(ps, ExtractPrintExprs(typesInfo, filename, line, offset, n, n.X)...)
		ps = append(ps, newPrintExpr(n.Index.Pos()-1-offset, n, original))
		ps = append(ps, ExtractPrintExprs(typesInfo, filename, line, offset, n, n.Index)...)
	case *ast.SelectorExpr:
		n := n.(*ast.SelectorExpr)
		ps = append(ps, ExtractPrintExprs(typesInfo, filename, line, offset, n, n.X)...)
		ps = append(ps, newPrintExpr(n.Sel.Pos()-offset, n, original))
	case *ast.CallExpr:
		n := n.(*ast.CallExpr)
		if IsBuiltinFunc(n) { // don't memorize buildin methods
			newExpr := CreateUntypedCallExprFromBuiltinCallExpr(n)

			ps = append(ps, newPrintExpr(n.Pos()-offset, newExpr, original))
			for _, arg := range n.Args {
				ps = append(ps, ExtractPrintExprs(typesInfo, filename, line, offset, n, arg)...)
			}

			*n = *newExpr
		} else if IsTypeConversion(typesInfo, n) {
			// T(v) can take only one argument.
			if len(n.Args) > 1 {
				panic("too many arguments for type conversion")
			} else if len(n.Args) < 1 {
				panic("missing argument for type conversion")
			}

			argsPrints := ExtractPrintExprs(typesInfo, filename, line, offset, n, n.Args[0])

			newExpr := CreateReflectInterfaceExpr(&ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   CreateReflectValueOfExpr(n.Args[0]),
					Sel: &ast.Ident{Name: "Convert"},
				},
				Args: []ast.Expr{
					CreateReflectTypeExprFromTypeExpr(n.Fun),
				},
			})

			ps = append(ps, newPrintExpr(n.Pos()-offset, newExpr, original))
			ps = append(ps, argsPrints...)

			*n = *newExpr
		} else {
			argsPrints := []printExpr{}
			for _, arg := range n.Args {
				argsPrints = append(argsPrints, ExtractPrintExprs(typesInfo, filename, line, offset, n, arg)...)
			}

			var memorized *ast.CallExpr

			if p, ok := parent.(*ast.CallExpr); ok && IsAssert(AssertImportIdent, p) {
				memorized = CreateMemorizedFuncCall(filename, line, n, "Bool")
			} else {
				memorized = CreateMemorizedFuncCall(filename, line, n, "Interface")
			}

			ps = append(ps, newPrintExpr(n.Pos()-offset, memorized, original))
			ps = append(ps, argsPrints...)

			*n = *memorized
		}
	case *ast.SliceExpr:
		n := n.(*ast.SliceExpr)
		ps = append(ps, ExtractPrintExprs(typesInfo, filename, line, offset, n, n.Low)...)
		ps = append(ps, ExtractPrintExprs(typesInfo, filename, line, offset, n, n.High)...)
		if n.Slice3 {
			ps = append(ps, ExtractPrintExprs(typesInfo, filename, line, offset, n, n.Max)...)
		}
	}

	return ps
}

// "a" -> 0
// "obj.Fn" -> 3
// "1 + 2" -> 2
// "(1)" -> 1
func ResultPosOf(n ast.Expr) token.Pos {
	switch n.(type) {
	case *ast.IndexExpr:
		n := n.(*ast.IndexExpr)
		return n.Index.Pos() - 1
	case *ast.SelectorExpr:
		n := n.(*ast.SelectorExpr)
		return ResultPosOf(n.Sel)
	case *ast.BinaryExpr:
		n := n.(*ast.BinaryExpr)
		return n.OpPos
	case *ast.ParenExpr:
		n := n.(*ast.ParenExpr)
		return ResultPosOf(n.X)
	default:
		return n.Pos()
	}
}
