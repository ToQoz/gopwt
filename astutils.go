package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"io"
	"strconv"
	"strings"
)

var (
	builtinFuncs = []string{
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

// replaceBinaryExpr replace oldExpr by newExpr in parent
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

// replaceAllRawStringLitByStringLit replaces all raw string literals in root by string literals.
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

// getAssertImport returns *ast.ImportSpec of "github.com/ToQoz/gopwt/assert"
// if it is not found, this returns nil
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

// isAssert returns ok if given CallExpr is github.com/ToQoz/gopwt/assert.OK or Require
func isAssert(x *ast.Ident, c *ast.CallExpr) bool {
	if s, ok := c.Fun.(*ast.SelectorExpr); ok {
		if xident, ok := s.X.(*ast.Ident); ok {
			return xident.Name == x.Name && (s.Sel.Name == "OK" || s.Sel.Name == "Require")
		}
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

func isEqualExpr(expr ast.Expr) bool {
	if b, ok := expr.(*ast.BinaryExpr); ok {
		return b.Op == token.EQL
	}

	return false
}

func isReflectDeepEqual(expr ast.Expr) bool {
	if c, ok := expr.(*ast.CallExpr); ok {
		if sel, ok := c.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok {
				return ident.Name == "reflect" && sel.Sel.Name == "DeepEqual"
			}
		}
	}

	return false
}

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

func createBoolIdent(v bool) *ast.Ident {
	var name string
	if v {
		name = "true"
	} else {
		name = "false"
	}
	return &ast.Ident{Name: name}
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
			createReflectValueOfExpr(n.Fun),
		},
	}

	args := []ast.Expr{}
	for _, a := range n.Args {
		args = append(args, createReflectValueOfExpr(a))
	}
	c.Args = append(c.Args, args...)

	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   translatedassertImportIdent,
			Sel: &ast.Ident{Name: "FRV" + returnType},
		},
		Args: []ast.Expr{c},
	}
}

// createReflectTypeExprFromTypeExpr create ast of reflect.Type from ast of type.
func createReflectTypeExprFromTypeExpr(t ast.Expr) ast.Expr {
	canUseCompositeLit := true

	if t, ok := t.(*ast.Ident); ok {
		switch t.Name {
		case "string", "rune",
			"uint", "uint8", "uint16", "uint32", "uint64",
			"int8", "int32", "int64", "int",
			"float32", "float64",
			"complex64", "complex128",
			"bool", "uintptr", "error":

			canUseCompositeLit = false
		}
	}

	if _, ok := t.(*ast.ChanType); ok {
		canUseCompositeLit = false
	}

	if !canUseCompositeLit {
		rv := &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: createReflectValueOfExpr(&ast.CallExpr{
					Fun:  &ast.Ident{Name: "new"},
					Args: []ast.Expr{t},
				}),
				Sel: &ast.Ident{Name: "Elem"},
			},
		}

		return &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   rv,
				Sel: &ast.Ident{Name: "Type"},
			},
		}
	}

	return createReflectTypeOfExpr(&ast.CompositeLit{Type: t})
}

func createReflectInterfaceExpr(rv ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   translatedassertImportIdent,
			Sel: &ast.Ident{Name: "RVInterface"},
		},
		Args: []ast.Expr{rv},
	}
}

func createReflectBoolExpr(rv ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   translatedassertImportIdent,
			Sel: &ast.Ident{Name: "RVBool"},
		},
		Args: []ast.Expr{rv},
	}
}

func createReflectValueOfExpr(v ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   translatedassertImportIdent,
			Sel: &ast.Ident{Name: "RVOf"},
		},
		Args: []ast.Expr{v},
	}
}

func createReflectTypeOfExpr(v ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   translatedassertImportIdent,
			Sel: &ast.Ident{Name: "RTOf"},
		},
		Args: []ast.Expr{v},
	}
}

func createPosValuePairExpr(ps []printExpr) []ast.Expr {
	args := []ast.Expr{}

	for _, n := range ps {
		powered := true
		if lit, ok := n.Expr.(*ast.BasicLit); ok {
			if lit.Kind == token.STRING {
				powered = len(strings.Split(lit.Value, "\\n")) > 1
			} else {
				powered = false
			}
		}
		if _, ok := n.Expr.(*ast.CompositeLit); ok {
			powered = false
		}

		a := &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   translatedassertImportIdent,
				Sel: &ast.Ident{Name: "NewPosValuePair"},
			},
			Args: []ast.Expr{
				&ast.BasicLit{Kind: token.INT, Value: strconv.Itoa(int(n.Pos))},
				n.Expr,
				createBoolIdent(powered),
				createRawStringLit(n.OriginalExpr),
			},
		}

		args = append(args, a)
	}

	return args
}

// VALUE:"foo"     ---> AST:`foo`
// VALUE:"foo`bar" ---> AST:`foo` + "`" + `bar` (because we can't escape ` in ``)
func createRawStringLit(s string) ast.Expr {
	segments := strings.Split(s, "`")

	if len(segments) == 1 {
		return &ast.BasicLit{Kind: token.STRING, Value: "`" + segments[0] + "`"}
	}

	rootBinary := &ast.BinaryExpr{
		Op: token.ADD,
	}
	binary := rootBinary
	for i, _ := range segments {
		// reverse for
		seg := segments[len(segments)-1-i]

		if i == 0 {
			binary.Y = &ast.BasicLit{Kind: token.STRING, Value: "`" + segments[len(segments)-1] + "`"}
		} else if i == len(segments)-1 {
			binary.X = &ast.BinaryExpr{
				X:  &ast.BasicLit{Kind: token.STRING, Value: "`" + seg + "`"},
				Y:  &ast.BasicLit{Kind: token.STRING, Value: `"` + "`" + `"`},
				Op: token.ADD,
			}
		} else {
			binary.X = &ast.BinaryExpr{
				X: &ast.BinaryExpr{
					Y:  &ast.BasicLit{Kind: token.STRING, Value: "`" + seg + "`"},
					Op: token.ADD,
				},
				Y:  &ast.BasicLit{Kind: token.STRING, Value: `"` + "`" + `"`},
				Op: token.ADD,
			}
			binary = binary.X.(*ast.BinaryExpr).X.(*ast.BinaryExpr)
		}
	}

	return rootBinary
}

func createArrayTypeCompositLit(typ string) *ast.CompositeLit {
	return &ast.CompositeLit{
		Type: &ast.ArrayType{
			Elt: &ast.Ident{
				Name: typ,
			},
		},
		Elts: []ast.Expr{},
	}
}

func sprintCode(n ast.Node) string {
	buf := bytes.NewBuffer([]byte{})
	fprintCode(buf, n)
	return buf.String()
}

func fprintCode(out io.Writer, n ast.Node) error {
	return printer.Fprint(out, token.NewFileSet(), n)
}

func inspectAssert(root ast.Node, fn func(*ast.CallExpr)) {
	ast.Inspect(root, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.CallExpr:
			n := n.(*ast.CallExpr)

			if _, ok := n.Fun.(*ast.FuncLit); ok {
				return true
			}

			if !isAssert(assertImportIdent, n) {
				// skip inspecting children in assert.OK
				return false
			}

			fn(n)
			return false
		}

		return true
	})
}
