package internal

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Rewrite rewrites assert to translatedassert in the given package
func (pkgCtx *PackageContext) RewritePackage() {
	wg := &sync.WaitGroup{}
	// Rewrite pkg files
	for _, file := range pkgCtx.GoFiles {
		wg.Add(1)
		go func(file *GoFile) {
			defer wg.Done()
			ctx := &Context{
				AssertImport:           &ast.Ident{Name: "assert"},
				TranslatedassertImport: &ast.Ident{Name: "translatedassert"},
			}

			out, err := os.OpenFile(file.Path, os.O_CREATE|os.O_WRONLY, file.Mode)
			if err != nil {
				pkgCtx.Error = err
				return
			}
			defer out.Close()

			if !IsTestGoFileName(file.Path) {
				_, err = io.Copy(out, file.Data)
				if err != nil {
					pkgCtx.Error = err
				}
				return
			}

			err = RewriteFile(pkgCtx, file, ctx, pkgCtx.TypeInfo, out)
			if err != nil {
				pkgCtx.Error = err
			}
			return
		}(file)
	}

	wg.Wait()
}

// RewriteFile rewrites assert to translatedassert in the given file
func RewriteFile(pkgCtx *PackageContext, file *GoFile, ctx *Context, typesInfo *types.Info, out io.Writer) error {
	gopwtMainDropped := DropGopwtEmpower(file.Normalized)

	assertImport := GetAssertImport(file.Normalized)
	if assertImport == nil {
		if !gopwtMainDropped {
			io.Copy(out, file.Data)
			return nil
		}
	} else {
		assertImport.Path.Value = `"github.com/ToQoz/gopwt/translatedassert"`

		if assertImport.Name != nil {
			ctx.AssertImport = assertImport.Name
			assertImport.Name = ctx.TranslatedassertImport
		}
	}

	assertPositions := []token.Position{}
	InspectAssert(ctx, file.Original, func(n *ast.CallExpr) {
		assertPositions = append(assertPositions, pkgCtx.OriginalFset.Position(n.Pos()))
	})
	i := 0
	InspectAssert(ctx, file.Normalized, func(n *ast.CallExpr) {
		pos := assertPositions[i]
		i++
		RewriteAssert(ctx, typesInfo, pos, n)
	})

	return printer.Fprint(out, pkgCtx.NormalizedFset, file.Normalized)
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
