package internal

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type PackageContext struct {
	Vendor          string
	HasVendor       bool
	OriginalFset    *token.FileSet
	NormalizedFset  *token.FileSet
	OriginalFiles   []*ast.File
	NormalizedFiles []*ast.File
	OutFiles        []*os.File
	Mode            []os.FileMode
}

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
	err := os.MkdirAll(filepath.Join(srcDir, importpath), os.ModePerm)
	if err != nil {
		return err
	}

	pkgCtx := &PackageContext{}
	pkgCtx.Vendor, pkgCtx.HasVendor = FindVendor(fpath, strings.Count(importpath, "/")+1)

	// Copy pkg/vendor to tempdir
	if pkgCtx.HasVendor {
		if err := CopyPackageVendor(pkgCtx.Vendor, fpath, importpath, srcDir); err != nil {
			return err
		}
	}

	// Copy pkg to tempdir
	err = CopyPackage(pkgCtx, fpath, importpath, srcDir)
	if err != nil {
		return err
	}

	// Check pkg types
	typesInfo, err := GetTypeInfo(pkgCtx, fpath, importpath, srcDir)
	if err != nil {
		return err
	}

	// Rewrite pkg
	return RewritePackage(pkgCtx, typesInfo, fpath, importpath, srcDir)
}

func CopyPackageVendor(vendor, pkgDir, importPath, tempGoSrcDir string) error {
	return filepath.Walk(vendor, func(path string, finfo os.FileInfo, err error) error {
		pathFromPkgDir, err := filepath.Rel(pkgDir, path)
		if err != nil {
			return err
		}

		outpath := filepath.Join(tempGoSrcDir, importPath, pathFromPkgDir)
		if finfo.IsDir() {
			return os.Mkdir(outpath, finfo.Mode())
		}
		out, err := os.OpenFile(outpath, os.O_WRONLY|os.O_CREATE, finfo.Mode())
		if err != nil {
			return err
		}
		defer out.Close()
		return CopyFile(path, out)
	})
}

func CopyPackage(pkgCtx *PackageContext, pkgDir string, importPath string, tempGoSrcDir string) (err error) {
	pkgCtx.OriginalFset = token.NewFileSet()
	pkgCtx.NormalizedFset = token.NewFileSet()

	err = filepath.Walk(pkgDir, func(path string, finfo os.FileInfo, err error) error {
		if path == pkgDir {
			return nil
		}

		if finfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil
		}

		pathFromPkgDir, err := filepath.Rel(pkgDir, path)
		if err != nil {
			return err
		}
		outpath := filepath.Join(tempGoSrcDir, importPath, pathFromPkgDir)

		if finfo.IsDir() {
			// copy all files in
			//   - <pkgDir>/testdata/**/*
			if IsTestdata(pathFromPkgDir) {
				return os.Mkdir(outpath, finfo.Mode())
			}

			return filepath.SkipDir
		}

		out, err := os.OpenFile(outpath, os.O_RDWR|os.O_CREATE, finfo.Mode())
		if err != nil {
			return err
		}
		// NOTE: Don't close
		// defer out.Close()

		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		// NOTE: Don't parse & rewrite gofiles under testdata. It is test data.
		if !IsGoFileName(path) || IsTestdata(pathFromPkgDir) {
			_, err := io.Copy(out, in)
			out.Close()
			return err
		}

		file, err := parser.ParseFile(pkgCtx.OriginalFset, path, in, 0)
		if err != nil {
			return err
		}
		pkgCtx.Mode = append(pkgCtx.Mode, finfo.Mode())
		pkgCtx.OriginalFiles = append(pkgCtx.OriginalFiles, file)

		defer func() {
			// Parse copied & normalized file. e.g. multi-line CompositeLit -> single-line
			out.Seek(0, SeekStart)
			// NOTE: printer.Fprint-ed output is should be valid
			pkgCtx.NormalizedFiles = append(pkgCtx.NormalizedFiles, MustParse(parser.ParseFile(pkgCtx.NormalizedFset, outpath, out, 0)))
			out.Seek(0, SeekStart)
			pkgCtx.OutFiles = append(pkgCtx.OutFiles, out)
		}()

		if !IsTestGoFileName(path) {
			return CopyFile(path, out)
		}

		ctx := &Context{
			AssertImport:           &ast.Ident{Name: "assert"},
			TranslatedassertImport: &ast.Ident{Name: "translatedassert"},
		}
		assertImport := GetAssertImport(file)
		if assertImport == nil {
			return CopyFile(path, out)
		}
		if assertImport.Name != nil {
			ctx.AssertImport = assertImport.Name
		}

		// Format and copy file
		//   - 1. replace rawStringLit to stringLit
		//   - 2. replace multiline compositLit to singleline one.
		//   - 3. copy
		InspectAssert(ctx, file, func(n *ast.CallExpr) {
			// 1. replace rawStringLit to stringLit
			ReplaceAllRawStringLitByStringLit(n)
		})

		// 2. replace multiline-compositLit by singleline-one by passing empty token.FileSet
		// 3. copy
		return printer.Fprint(out, token.NewFileSet(), file)
	})

	return
}

func RewritePackage(pkgCtx *PackageContext, typesInfo *types.Info, pkgDir, importPath string, tempGoSrcDir string) error {
	// Rewrite files
	Assert(len(pkgCtx.NormalizedFiles) == len(pkgCtx.OriginalFiles), "normalized files count should be equals to original files count.")
	for i := 0; i < len(pkgCtx.NormalizedFiles); i++ {
		err := func() error {
			defer pkgCtx.OutFiles[i].Close()

			ctx := &Context{
				AssertImport:           &ast.Ident{Name: "assert"},
				TranslatedassertImport: &ast.Ident{Name: "translatedassert"},
			}

			path := pkgCtx.NormalizedFset.File(pkgCtx.NormalizedFiles[i].Package).Name()

			if !IsTestGoFileName(path) {
				return nil
			}

			gopwtMainDropped := DropGopwtEmpower(pkgCtx.NormalizedFiles[i])

			assertImport := GetAssertImport(pkgCtx.NormalizedFiles[i])
			if assertImport == nil {
				if !gopwtMainDropped {
					return nil
				}
			} else {
				assertImport.Path.Value = `"github.com/ToQoz/gopwt/translatedassert"`

				if assertImport.Name != nil {
					ctx.AssertImport = assertImport.Name
					assertImport.Name = ctx.TranslatedassertImport
				}
			}

			pkgCtx.OutFiles[i].Truncate(0)
			return RewriteFile(pkgCtx, i, ctx, typesInfo, pkgCtx.OutFiles[i])
		}()
		if err != nil {
			return err
		}
	}

	return nil
}

func RewriteFile(pkgCtx *PackageContext, fileIndex int, ctx *Context, typesInfo *types.Info, out io.Writer) error {
	assertPositions := []token.Position{}
	InspectAssert(ctx, pkgCtx.OriginalFiles[fileIndex], func(n *ast.CallExpr) {
		assertPositions = append(assertPositions, pkgCtx.OriginalFset.Position(n.Pos()))
	})
	i := 0
	InspectAssert(ctx, pkgCtx.NormalizedFiles[fileIndex], func(n *ast.CallExpr) {
		pos := assertPositions[i]
		i++
		RewriteAssert(ctx, typesInfo, pos, n)
	})

	return printer.Fprint(out, pkgCtx.NormalizedFset, pkgCtx.NormalizedFiles[fileIndex])
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
	// assert.OK -> translatedassert.OK
	if sel, ok := n.Fun.(*ast.SelectorExpr); ok {
		sel.X = ctx.TranslatedassertImport
	}
	// OK -> translatedassert.OK
	if name, ok := n.Fun.(*ast.Ident); ok {
		n.Fun = &ast.SelectorExpr{
			X:   ctx.TranslatedassertImport,
			Sel: name,
		}
	}
}

type printExpr struct {
	Pos          int
	Expr         ast.Expr
	OriginalExpr string
}

func NewPrintExpr(pos token.Pos, newExpr ast.Expr, originalExpr string) printExpr {
	return printExpr{Pos: int(pos), Expr: newExpr, OriginalExpr: originalExpr}
}

func ExtractPrintExprs(ctx *Context, typesInfo *types.Info, filename string, line int, offset token.Pos, parent ast.Expr, n ast.Expr) []printExpr {
	ps := []printExpr{}

	original := SprintCode(n)

	switch n.(type) {
	case *ast.BasicLit:
		n := n.(*ast.BasicLit)
		ps = append(ps, NewPrintExpr(n.Pos()-offset, n, original))
	case *ast.CompositeLit:
		n := n.(*ast.CompositeLit)

		ps = append(ps, NewPrintExpr(n.Pos()-offset, n, original))
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
		ps = append(ps, NewPrintExpr(n.Pos()-offset, n, original))
	case *ast.ParenExpr:
		n := n.(*ast.ParenExpr)
		ps = append(ps, ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.X)...)
	case *ast.StarExpr:
		n := n.(*ast.StarExpr)
		ps = append(ps, NewPrintExpr(n.Pos()-offset, n, original))
		ps = append(ps, ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.X)...)
	// "+" | "-" | "!" | "^" | "*" | "&" | "<-" .
	case *ast.UnaryExpr:
		n := n.(*ast.UnaryExpr)

		x := ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.X)

		untyped, ok := CreateUntypedExprFromUnaryExpr(ctx, n).(*ast.CallExpr)
		if ok {
			newExpr := CreateMemorizedFuncCall(ctx, filename, line, n.Pos(), untyped, "Interface")
			ReplaceUnaryExpr(parent, n, newExpr)
			ps = append(ps, NewPrintExpr(n.Pos()-offset, newExpr, original))
		} else {
			ps = append(ps, NewPrintExpr(n.Pos()-offset, n, original))
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
		ps = append(ps, NewPrintExpr(n.OpPos-offset, newExpr, original))
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
		ps = append(ps, NewPrintExpr(n.Index.Pos()-1-offset, n, original))
		ps = append(ps, ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.Index)...)
	case *ast.SelectorExpr:
		n := n.(*ast.SelectorExpr)
		ps = append(ps, ExtractPrintExprs(ctx, typesInfo, filename, line, offset, n, n.X)...)
		ps = append(ps, NewPrintExpr(n.Sel.Pos()-offset, n, original))
	case *ast.CallExpr:
		n := n.(*ast.CallExpr)
		if IsBuiltinFunc(n) { // don't memorize buildin methods
			newExpr := CreateUntypedCallExprFromBuiltinCallExpr(ctx, n)

			ps = append(ps, NewPrintExpr(n.Pos()-offset, newExpr, original))
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

			ps = append(ps, NewPrintExpr(n.Pos()-offset, newExpr, original))
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

			ps = append(ps, NewPrintExpr(n.Pos()-offset, memorized, original))
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
