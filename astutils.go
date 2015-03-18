package main

import (
	"fmt"
	"go/ast"
	"go/token"
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

// isAssert returns ok if given CallExpr is github.com/ToQoz/gopwt/assert.OK
func isAssert(x *ast.Ident, c *ast.CallExpr) bool {
	if s, ok := c.Fun.(*ast.SelectorExpr); ok {
		if xident, ok := s.X.(*ast.Ident); ok {
			return xident.Name == x.Name && s.Sel.Name == "OK"
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
