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
	"sync"
)

type Context struct {
	TranslatedassertImport *ast.Ident
	AssertImport           *ast.Ident
}

var (
	Testdata   = "testdata"
	TermWidth  = 0
	WorkingDir = ""
	Verbose    = false
)

func Rewrite(gopath string, importpath, fpath string) error {
	srcDir := filepath.Join(gopath, "src")
	testdata := strings.Split(Testdata, ",")

	err := filepath.Walk(fpath, func(path string, fInfo os.FileInfo, err error) error {
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

		rel, err := filepath.Rel(fpath, path)
		if err != nil {
			return err
		}

		for _, tdata := range testdata {
			if strings.Split(rel, string(filepath.Separator))[0] == tdata {
				return filepath.SkipDir
			}
		}

		if rel != "." {
			return filepath.SkipDir
		}

		importpath := filepath.Join(importpath, rel)

		err = os.MkdirAll(filepath.Join(srcDir, importpath), os.ModePerm)
		if err != nil {
			return err
		}

		rewritePackage(path, importpath, srcDir)
		return nil
	})

	return err
}

type copyTarget struct {
	path    string
	outPath string
	mode    os.FileMode
}

type file struct {
	original *ast.File
	rewrited *ast.File
	mode     os.FileMode
}

func rewritePackage(pkgDir, importPath string, tempGoSrcDir string) error {
	// Copy to tempdir
	originalFset, fset, files, err := copyPackage(pkgDir, importPath, tempGoSrcDir)
	if err != nil {
		return err
	}

	rewritedFiles := []*ast.File{}
	for _, files := range files {
		rewritedFiles = append(rewritedFiles, files.rewrited)
	}

	typesInfo, err := GetTypeInfo(pkgDir, importPath, tempGoSrcDir, fset, rewritedFiles)
	if err != nil {
		return err
	}

	// Rewrite files
	for _, f := range files {
		path := fset.File(f.rewrited.Package).Name()

		if !IsTestGoFileName(path) {
			continue
		}

		gopwtMainDropped := DropGopwtEmpower(f.rewrited)

		ctx := &Context{
			AssertImport:           &ast.Ident{Name: "assert"},
			TranslatedassertImport: &ast.Ident{Name: "translatedassertImport"},
		}
		assertImport := GetAssertImport(f.rewrited)
		if assertImport == nil {
			if !gopwtMainDropped {
				continue
			}
		} else {
			assertImport.Path.Value = `"github.com/ToQoz/gopwt/translatedassert"`

			if assertImport.Name != nil {
				ctx.AssertImport = assertImport.Name
				assertImport.Name = ctx.TranslatedassertImport
			}
		}

		out, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, f.mode)
		if err != nil {
			return err
		}
		defer out.Close()
		return RewriteFile(ctx, typesInfo, fset, originalFset, f.rewrited, f.original, out)
	}

	return nil
}

func copyPackage(pkgDir, importPath string, tempGoSrcDir string) (originalFset *token.FileSet, fset *token.FileSet, files []*file, err error) {
	testdata := strings.Split(Testdata, ",")
	originalFset = token.NewFileSet()
	targets := []copyTarget{}

	// Collect fset(*token.FileSet) and files([]*ast.File)
	fset = token.NewFileSet()
	files = []*file{}

	err = filepath.Walk(pkgDir, func(path string, fInfo os.FileInfo, err error) error {
		mode := fInfo.Mode()
		if mode&os.ModeSymlink == os.ModeSymlink {
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
			for _, tdata := range testdata {
				if strings.Split(pathFromImportDir, string(filepath.Separator))[0] == tdata {
					di, err := os.Stat(filepath.Dir(path))
					if err != nil {
						return err
					}
					return os.Mkdir(outPath, di.Mode())
				}
			}

			return filepath.SkipDir
		}

		targets = append(targets, copyTarget{path: path, mode: mode, outPath: outPath})
		return nil
	})

	fn := func(path string, outPath string, mode os.FileMode, wg *sync.WaitGroup) error {
		defer wg.Done()

		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		out, err := os.OpenFile(outPath, os.O_RDWR|os.O_CREATE, mode)
		if err != nil {
			return err
		}

		if !IsGoFileName(path) {
			_, err = io.Copy(out, in)
			out.Close()
			return err
		}

		// parse only pkg's GoFiles
		if filepath.Dir(path) != pkgDir {
			_, err = io.Copy(out, in)
			out.Close()
			return err
		}

		a, err := parser.ParseFile(originalFset, path, in, 0)
		if err != nil {
			return err
		}
		origA := a
		in.Seek(0, os.SEEK_SET) // NOTE: go1.5 and go1.6 doesn't have io.SeekStart

		defer func() {
			// parse copyed & normalized file. e.g. multi-line CompositeLit -> single-line
			out.Seek(0, 0)
			a, err := parser.ParseFile(fset, outPath, out, 0)
			if err != nil {
				// The file that printed by printer.Fprint code must be parsable
				panic(err)
			}
			f := &file{
				rewrited: a,
				original: origA,
				mode:     mode,
			}
			files = append(files, f)
			out.Close()
		}()

		if !IsTestGoFileName(path) {
			_, err = io.Copy(out, in)
			return err
		}

		ctx := &Context{
			AssertImport:           &ast.Ident{Name: "assert"},
			TranslatedassertImport: &ast.Ident{Name: "translatedassertImport"},
		}
		assertImport := GetAssertImport(a)
		if assertImport == nil {
			_, err = io.Copy(out, in)
			return err
		}
		if assertImport.Name != nil {
			ctx.AssertImport = assertImport.Name
		}

		// Format and copy file
		//   - 1. replace rawStringLit to stringLit
		//   - 2. replace multiline compositLit to singleline one.
		//   - 3. copy
		InspectAssert(ctx, a, func(n *ast.CallExpr) {
			// 1. replace rawStringLit to stringLit
			ReplaceAllRawStringLitByStringLit(n)
		})

		// 2. replace multiline-compositLit by singleline-one by passing empty token.FileSet
		// 3. copy
		return printer.Fprint(out, token.NewFileSet(), a)
	}
	wg := &sync.WaitGroup{}
	for _, t := range targets {
		wg.Add(1)
		go fn(t.path, t.outPath, t.mode, wg)
	}
	wg.Wait()
	return
}

func RewriteFile(ctx *Context, typesInfo *types.Info, fset, originalFset *token.FileSet, file, origFile *ast.File, out io.Writer) error {
	assertPositions := []token.Position{}
	InspectAssert(ctx, origFile, func(n *ast.CallExpr) {
		assertPositions = append(assertPositions, originalFset.Position(n.Pos()))
	})
	i := 0
	InspectAssert(ctx, file, func(n *ast.CallExpr) {
		pos := assertPositions[i]
		i++
		RewriteAssert(ctx, typesInfo, pos, n)
	})

	return printer.Fprint(out, fset, file)
}

// RewriteAssert rewrites assert to translatedassert
func RewriteAssert(ctx *Context, typesInfo *types.Info, position token.Position, n *ast.CallExpr) {
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
	fname := position.Filename
	if strings.HasPrefix(fname, WorkingDir) {
		fname = strings.TrimPrefix(fname, WorkingDir)
	}
	n.Args = append(n.Args, CreateRawStringLit(fname))
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
	extractedPrintExprs := ExtractPrintExprs(ctx, typesInfo, position.Filename, position.Line, posOffset, n, n.Args[1])
	n.Args = append(n.Args, CreatePosValuePairExpr(ctx, extractedPrintExprs)...)
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

func ExtractPrintExprs(ctx *Context, typesInfo *types.Info, filename string, line int, offset token.Pos, parent ast.Expr, n ast.Expr) []printExpr {
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
			ps = append(ps, ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, elt)...)
		}
	case *ast.KeyValueExpr:
		n := n.(*ast.KeyValueExpr)

		if IsMapType(parent) {
			ps = append(ps, ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.Key)...)
		}

		ps = append(ps, ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.Value)...)
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
		ps = append(ps, ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.X)...)
	case *ast.StarExpr:
		n := n.(*ast.StarExpr)
		ps = append(ps, newPrintExpr(n.Pos()-offset, n, original))
		ps = append(ps, ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.X)...)
	// "+" | "-" | "!" | "^" | "*" | "&" | "<-" .
	case *ast.UnaryExpr:
		n := n.(*ast.UnaryExpr)

		x := ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.X)

		untyped, ok := CreateUntypedExprFromUnaryExpr(ctx, n).(*ast.CallExpr)
		if ok {
			newExpr := CreateMemorizedFuncCall(ctx, filename, line, n.Pos(), untyped, "Interface")
			ReplaceUnaryExpr(parent, n, newExpr)
			ps = append(ps, newPrintExpr(n.Pos()-offset, newExpr, original))
		} else {
			ps = append(ps, newPrintExpr(n.Pos()-offset, n, original))
		}

		ps = append(ps, x...)
	case *ast.BinaryExpr:
		n := n.(*ast.BinaryExpr)

		var x, y []printExpr

		x = ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.X)
		y = ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.Y)

		newExpr := CreateUntypedExprFromBinaryExpr(ctx, n)
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
		ps = append(ps, ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.X)...)
		ps = append(ps, newPrintExpr(n.Index.Pos()-1-offset, n, original))
		ps = append(ps, ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.Index)...)
	case *ast.SelectorExpr:
		n := n.(*ast.SelectorExpr)
		ps = append(ps, ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.X)...)
		ps = append(ps, newPrintExpr(n.Sel.Pos()-offset, n, original))
	case *ast.CallExpr:
		n := n.(*ast.CallExpr)
		if IsBuiltinFunc(n) { // don't memorize buildin methods
			newExpr := CreateUntypedCallExprFromBuiltinCallExpr(ctx, n)

			ps = append(ps, newPrintExpr(n.Pos()-offset, newExpr, original))
			for _, arg := range n.Args {
				ps = append(ps, ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, arg)...)
			}

			*n = *newExpr
		} else if IsTypeConversion(typesInfo, n) {
			// T(v) can take only one argument.
			if len(n.Args) > 1 {
				panic("too many arguments for type conversion")
			} else if len(n.Args) < 1 {
				panic("missing argument for type conversion")
			}

			argsPrints := ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.Args[0])

			newExpr := CreateReflectInterfaceExpr(ctx, &ast.CallExpr{
				Fun: &ast.SelectorExpr{
					X:   CreateReflectValueOfExpr(ctx, n.Args[0]),
					Sel: &ast.Ident{Name: "Convert"},
				},
				Args: []ast.Expr{
					CreateReflectTypeExprFromTypeExpr(ctx, n.Fun),
				},
			})

			ps = append(ps, newPrintExpr(n.Pos()-offset, newExpr, original))
			ps = append(ps, argsPrints...)

			*n = *newExpr
		} else {
			argsPrints := []printExpr{}
			for _, arg := range n.Args {
				argsPrints = append(argsPrints, ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, arg)...)
			}

			var memorized *ast.CallExpr

			if p, ok := parent.(*ast.CallExpr); ok && IsAssert(ctx.AssertImport, p) {
				memorized = CreateMemorizedFuncCall(ctx, filename, line, n.Pos(), n, "Bool")
			} else {
				memorized = CreateMemorizedFuncCall(ctx, filename, line, n.Pos(), n, "Interface")
			}

			ps = append(ps, newPrintExpr(n.Pos()-offset, memorized, original))
			ps = append(ps, argsPrints...)

			*n = *memorized
		}
	case *ast.SliceExpr:
		n := n.(*ast.SliceExpr)
		ps = append(ps, ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.Low)...)
		ps = append(ps, ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.High)...)
		if n.Slice3 {
			ps = append(ps, ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.Max)...)
		}
	}

	return ps
}

// ResultPosOf returns result position of given expr.
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
