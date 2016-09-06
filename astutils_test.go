package gopwt

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/ToQoz/gopwt/assert"
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

func TestInspectAssert(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "./testdata/inspect_assert_tests/main.go", nil, 0)
	assert.Require(t, err == nil)

	asserts := []string{}
	inspectAssert(file, func(n *ast.CallExpr) {
		asserts = append(asserts, astToCode(n))
	})
	assert.OK(t, len(asserts) == 2)
	assert.OK(t, asserts[0] == `assert.OK(nil, "basic" == "basic")`)
	assert.OK(t, asserts[1] == `assert.OK(nil, "in func" == "in func")`)
}

func TestGetAssertImport(t *testing.T) {
	var f *ast.File
	var err error
	var importSpec *ast.ImportSpec

	// default import
	f, err = parser.ParseFile(token.NewFileSet(), "./testdata/get_asert_import_tests/default_import.go", nil, 0)
	assert.Require(t, err == nil)

	importSpec = getAssertImport(f)
	assert.OK(t, importSpec.Name == nil)
	assert.OK(t, importSpec.Path.Value == `"github.com/ToQoz/gopwt/assert"`)

	// named import
	f, err = parser.ParseFile(token.NewFileSet(), "./testdata/get_asert_import_tests/named_import.go", nil, 0)
	assert.Require(t, err == nil)

	importSpec = getAssertImport(f)
	assert.OK(t, importSpec.Name.Name == `powerAssert`)
	assert.OK(t, importSpec.Path.Value == `"github.com/ToQoz/gopwt/assert"`)
}

func TestReplaceBinaryExpr(t *testing.T) {
	var parent ast.Expr
	var bin *ast.BinaryExpr

	parent = mustParseExpr("a(b+c)")
	bin = parent.(*ast.CallExpr).Args[0].(*ast.BinaryExpr)
	replaceBinaryExpr(parent, bin, mustParseExpr("x(1, 2)"))
	assert.OK(t, astToCode(parent) == "a(x(1, 2))")

	parent = mustParseExpr("struct{a int}{a: 1+2}").(*ast.CompositeLit).Elts[0]
	bin = parent.(*ast.KeyValueExpr).Value.(*ast.BinaryExpr)
	replaceBinaryExpr(parent, bin, mustParseExpr("x(1, 2)"))
	assert.OK(t, astToCode(parent) == "a: x(1, 2)")

	parent = mustParseExpr("a[1+2]")
	bin = parent.(*ast.IndexExpr).Index.(*ast.BinaryExpr)
	replaceBinaryExpr(parent, bin, mustParseExpr("x(1, 2)"))
	assert.OK(t, astToCode(parent) == "a[x(1, 2)]")

	parent = mustParseExpr("(1+2)")
	bin = parent.(*ast.ParenExpr).X.(*ast.BinaryExpr)
	replaceBinaryExpr(parent, bin, mustParseExpr("x(1, 2)"))
	assert.OK(t, astToCode(parent) == "(x(1, 2))")

	parent = mustParseExpr("1+1+2")
	bin = parent.(*ast.BinaryExpr).X.(*ast.BinaryExpr)
	replaceBinaryExpr(parent, bin, mustParseExpr("x(1, 2)"))
	assert.OK(t, astToCode(parent) == "x(1, 2) + 2")
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
