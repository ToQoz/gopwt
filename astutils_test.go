package main

import (
	"github.com/ToQoz/gopwt/assert"
	"go/ast"
	"go/parser"
	"strings"
	"testing"
)

func TestIsAssert_Regression(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			t.Error(p)
		}
	}()

	exprs, err := parser.ParseExpr(`(&BrowseNode{Path: "/foo/bar/"}).ParentPath()`)
	if err != nil {
		panic(err)
	}

	// Check no panic on isAssert when *ast.SelectorExpr.X is not *ast.Ident
	isAssert(assertImportIdent, exprs.(*ast.CallExpr))
}

func TestReplaceAllRawStringLitByStringLit(t *testing.T) {
	n := mustParseExpr(`func() string {
		return ` + "`" + `raw"string` + "`" + `
		}`)
	replaceAllRawStringLitByStringLit(n)
	assert.OK(t, astToCode(n) == `func() string {
	return "raw\"string"
}`)
}

func TestCreateRawStringLit(t *testing.T) {
	bq := "\"`\"" // back quote string literal "`"
	assert.OK(t, astToCode(createRawStringLit("foo")) == "`foo`")
	assert.OK(t, astToCode(createRawStringLit("foo`bar")) == strings.Join([]string{"`foo`", "`bar`"}, " + "+bq+" + "))
	assert.OK(t, astToCode(createRawStringLit("f`o`o")) == strings.Join([]string{"`f`", "`o`", "`o`"}, " + "+bq+" + "))
	assert.OK(t, astToCode(createRawStringLit("`ba`ba`ba`")) == strings.Join([]string{"``", "`ba`", "`ba`", "`ba`", "``"}, " + "+bq+" + "))
}

func TestCreateUntypedCallExprFromBuiltinCallExpr(t *testing.T) {
	var expr ast.Expr

	expr = createUntypedCallExprFromBuiltinCallExpr(mustParseExpr("make(typ)").(*ast.CallExpr))
	assert.OK(t, astToCode(expr) == "translatedassert.Bmake(translatedassert.RTOf(typ{}))")

	expr = createUntypedCallExprFromBuiltinCallExpr(mustParseExpr("new(typ)").(*ast.CallExpr))
	assert.OK(t, astToCode(expr) == "translatedassert.Bnew(translatedassert.RTOf(typ{}))")

	expr = createUntypedCallExprFromBuiltinCallExpr(mustParseExpr("cap(a)").(*ast.CallExpr))
	assert.OK(t, astToCode(expr) == "translatedassert.Bcap(a)")

	expr = createUntypedCallExprFromBuiltinCallExpr(mustParseExpr("complex(a, b)").(*ast.CallExpr))
	assert.OK(t, astToCode(expr) == "translatedassert.Bcomplex(a, b)")

	expr = createUntypedCallExprFromBuiltinCallExpr(mustParseExpr("copy(a, b)").(*ast.CallExpr))
	assert.OK(t, astToCode(expr) == "translatedassert.Bcopy(a, b)")

	expr = createUntypedCallExprFromBuiltinCallExpr(mustParseExpr("imag(a)").(*ast.CallExpr))
	assert.OK(t, astToCode(expr) == "translatedassert.Bimag(a)")

	expr = createUntypedCallExprFromBuiltinCallExpr(mustParseExpr("len(a)").(*ast.CallExpr))
	assert.OK(t, astToCode(expr) == "translatedassert.Blen(a)")

	expr = createUntypedCallExprFromBuiltinCallExpr(mustParseExpr("real(a)").(*ast.CallExpr))
	assert.OK(t, astToCode(expr) == "translatedassert.Breal(a)")
}

func TestCreateUntypedExprFromBinaryExpr(t *testing.T) {
	var f *ast.CallExpr

	f = createUntypedExprFromBinaryExpr(mustParseExpr("1 + 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "OpADD")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "2")

	f = createUntypedExprFromBinaryExpr(mustParseExpr("1 - 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "OpSUB")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "2")

	f = createUntypedExprFromBinaryExpr(mustParseExpr("1 * 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "OpMUL")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "2")

	f = createUntypedExprFromBinaryExpr(mustParseExpr("1 / 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "OpQUO")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "2")

	f = createUntypedExprFromBinaryExpr(mustParseExpr("1 % 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "OpREM")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "2")

	f = createUntypedExprFromBinaryExpr(mustParseExpr("1 & 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "OpAND")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "2")

	f = createUntypedExprFromBinaryExpr(mustParseExpr("1 | 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "OpOR")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "2")

	f = createUntypedExprFromBinaryExpr(mustParseExpr("1 ^ 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "OpXOR")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "2")

	f = createUntypedExprFromBinaryExpr(mustParseExpr("1 &^ 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "OpANDNOT")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "2")

	f = createUntypedExprFromBinaryExpr(mustParseExpr("1 << 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "OpSHL")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "2")

	f = createUntypedExprFromBinaryExpr(mustParseExpr("1 >> 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "OpSHR")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "2")

	f = createUntypedExprFromBinaryExpr(mustParseExpr("true && true").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "OpLAND")
	assert.OK(t, f.Args[0].(*ast.Ident).Name == "true")
	assert.OK(t, f.Args[1].(*ast.Ident).Name == "true")

	f = createUntypedExprFromBinaryExpr(mustParseExpr("false || false").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "OpLOR")
	assert.OK(t, f.Args[0].(*ast.Ident).Name == "false")
	assert.OK(t, f.Args[1].(*ast.Ident).Name == "false")
}
