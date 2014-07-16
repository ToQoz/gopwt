package main

import (
	"github.com/ToQoz/gopwt/assert"
	"go/ast"
	"go/parser"
	"testing"
)

func TestExtractPrintExprs_SingleLineStringLit(t *testing.T) {
	ps := extractPrintExprs(nil, mustParseExpr(`"foo" == "bar"`))
	assert.OK(t, len(ps) == 1)
	assert.OK(t, ps[0].Pos == len(`"foo" `)+1)
}

func TestExtractPrintExprs_MultiLineStringLit(t *testing.T) {
	ps := extractPrintExprs(nil, mustParseExpr(`"foo\nbar" == "bar"`))
	assert.OK(t, len(ps) == 2)
	assert.OK(t, ps[0].Pos == 1)
	assert.OK(t, ps[0].Expr.(*ast.BasicLit).Value == `"foo\nbar"`)
}

func TestExtractPrintExprs_UnaryExpr(t *testing.T) {
	ps := extractPrintExprs(nil, mustParseExpr("!a"))
	assert.OK(t, len(ps) == 2)
	assert.OK(t, ps[0].Pos == 1)
	assert.OK(t, ps[0].Expr.(*ast.UnaryExpr).X.(*ast.Ident).Name == "a")
	assert.OK(t, ps[1].Pos == 2)
	assert.OK(t, ps[1].Expr.(*ast.Ident).Name == "a")
}

func TestExtractPrintExprs_StarExpr(t *testing.T) {
	ps := extractPrintExprs(nil, mustParseExpr("*a"))
	assert.OK(t, len(ps) == 2)
	assert.OK(t, ps[0].Pos == 1)
	assert.OK(t, ps[0].Expr.(*ast.StarExpr).X.(*ast.Ident).Name == "a")
	assert.OK(t, ps[1].Pos == 2)
	assert.OK(t, ps[1].Expr.(*ast.Ident).Name == "a")
}

func TestExtractPrintExprs_IndexExpr(t *testing.T) {
	ps := extractPrintExprs(nil, mustParseExpr("ary[i] == ary2[i2]"))
	assert.OK(t, len(ps) == 5)
	assert.OK(t, ps[0].Pos == 1)
	assert.OK(t, ps[0].Expr.(*ast.Ident).Name == "ary")
	assert.OK(t, ps[1].Pos == len("ary[")+1)
	assert.OK(t, ps[1].Expr.(*ast.Ident).Name == "i")
	assert.OK(t, ps[2].Pos == len("ary[i] ")+1)
	assert.OK(t, ps[3].Pos == len("ary[i] == ")+1)
	assert.OK(t, ps[3].Expr.(*ast.Ident).Name == "ary2")
	assert.OK(t, ps[4].Pos == len("ary[i] == ary2[")+1)
	assert.OK(t, ps[4].Expr.(*ast.Ident).Name == "i2")
}

func TestExtractPrintExprs_ArrayType(t *testing.T) {
	ps := extractPrintExprs(nil, mustParseExpr("reflect.DeepEqual([]string{c}, []string{})"))
	assert.OK(t, len(ps) == 2)
	assert.OK(t, ps[0].Pos == 1)
	assert.OK(t, ps[0].Expr.(*ast.CallExpr).Fun.(*ast.SelectorExpr).X.(*ast.Ident).Name == "reflect")
	assert.OK(t, ps[0].Expr.(*ast.CallExpr).Fun.(*ast.SelectorExpr).Sel.Name == "DeepEqual")
	assert.OK(t, ps[1].Pos == len("reflect.DeepEqual([]string{")+1)
	assert.OK(t, ps[1].Expr.(*ast.Ident).Name == "c")

	ps = extractPrintExprs(nil, mustParseExpr("reflect.DeepEqual([4]string{d}, []string{})"))
	assert.OK(t, len(ps) == 2)
	assert.OK(t, ps[0].Pos == 1)
	assert.OK(t, ps[0].Expr.(*ast.CallExpr).Fun.(*ast.SelectorExpr).X.(*ast.Ident).Name == "reflect")
	assert.OK(t, ps[0].Expr.(*ast.CallExpr).Fun.(*ast.SelectorExpr).Sel.Name == "DeepEqual")
	assert.OK(t, ps[1].Pos == len("reflect.DeepEqual([4]string{")+1)
	assert.OK(t, ps[1].Expr.(*ast.Ident).Name == "d")
}

func TestExtractPrintExprs_MapType(t *testing.T) {
	ps := extractPrintExprs(nil, mustParseExpr("reflect.DeepEqual(map[string]string{a:b}, map[string]string{})"))
	assert.OK(t, len(ps) == 3)
	assert.OK(t, ps[0].Pos == 1)
	assert.OK(t, ps[0].Expr.(*ast.CallExpr).Fun.(*ast.SelectorExpr).X.(*ast.Ident).Name == "reflect")
	assert.OK(t, ps[0].Expr.(*ast.CallExpr).Fun.(*ast.SelectorExpr).Sel.Name == "DeepEqual")
	assert.OK(t, ps[1].Pos == len("reflect.DeepEqual(map[string]string{")+1)
	assert.OK(t, ps[1].Expr.(*ast.Ident).Name == "a")
	assert.OK(t, ps[2].Pos == len("reflect.DeepEqual(map[string]string{a:")+1)
	assert.OK(t, ps[2].Expr.(*ast.Ident).Name == "b")
}

func TestExtractPrintExprs_StructType(t *testing.T) {
	ps := extractPrintExprs(nil, mustParseExpr("reflect.DeepEqual(struct{Name string}{}, struct{Name string}{Name: foo})"))
	assert.OK(t, len(ps) == 2)
	assert.OK(t, ps[0].Pos == 1)
	assert.OK(t, ps[0].Expr.(*ast.CallExpr).Fun.(*ast.SelectorExpr).X.(*ast.Ident).Name == "reflect")
	assert.OK(t, ps[0].Expr.(*ast.CallExpr).Fun.(*ast.SelectorExpr).Sel.Name == "DeepEqual")
	assert.OK(t, ps[1].Pos == len("reflect.DeepEqual(struct{Name string}{}, struct{Name string}{Name: ")+1)
	assert.OK(t, ps[1].Expr.(*ast.Ident).Name == "foo")
}

func mustParseExpr(s string) ast.Expr {
	e, err := parser.ParseExpr(s)
	if err != nil {
		panic(err)
	}

	return e
}
