package internal_test

import (
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/ToQoz/gopwt/assert"
	. "github.com/ToQoz/gopwt/translator/internal"
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

	// Check no panic on IsAssert when *ast.SelectorExpr.X is not *ast.Ident
	IsAssert(ctx.AssertImport, exprs.(*ast.CallExpr))
}

func TestInspectAssert(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "./testdata/inspect_assert_tests/main.go", nil, 0)
	assert.Require(t, err == nil)

	asserts := []string{}
	InspectAssert(ctx, file, func(n *ast.CallExpr) {
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

	importSpec = GetAssertImport(f)
	assert.OK(t, importSpec.Name == nil)
	assert.OK(t, importSpec.Path.Value == `"github.com/ToQoz/gopwt/assert"`)

	// named import
	f, err = parser.ParseFile(token.NewFileSet(), "./testdata/get_asert_import_tests/named_import.go", nil, 0)
	assert.Require(t, err == nil)

	importSpec = GetAssertImport(f)
	assert.OK(t, importSpec.Name.Name == `powerAssert`)
	assert.OK(t, importSpec.Path.Value == `"github.com/ToQoz/gopwt/assert"`)
}

func TestDropGopwtEmpower(t *testing.T) {
	test := func(input, expected string) {
		tmp, err := ioutil.TempFile("", "gopwt_TestDropGopwtEmpower")
		assert.Require(t, err == nil)

		defer os.Remove(tmp.Name())
		defer tmp.Close()

		_, err = tmp.Write([]byte(input))
		assert.Require(t, err == nil)

		fset := token.NewFileSet()

		f, err := parser.ParseFile(fset, tmp.Name(), nil, 0)
		assert.Require(t, err == nil)
		DropGopwtEmpower(f)

		tmp.Seek(0, 0)
		tmp.Truncate(0)

		err = printer.Fprint(tmp, fset, f)
		assert.Require(t, err == nil)

		data, err := ioutil.ReadFile(tmp.Name())
		assert.Require(t, err == nil)
		assert.OK(t, string(data) == expected)
	}

	test(`package apkg

import (
	"testing"
	"os"

	"github.com/ToQoz/gopwt"
)

func TestMain(m *testing.M) {
	gopwt.Empower()
	os.Exit(m.Run())
}
`, `package apkg

import (
	"testing"
	"os"
)

func TestMain(m *testing.M) {

	os.Exit(m.Run())
}
`)

	test(`package apkg

import (
	"testing"
	"os"

	gopwt2 "github.com/ToQoz/gopwt"
)

func TestMain(m *testing.M) {
	gopwt2.Empower()
	os.Exit(m.Run())
}
`, `package apkg

import (
	"testing"
	"os"
)

func TestMain(m *testing.M) {

	os.Exit(m.Run())
}
`)

	test(`package gopwt

import (
	"testing"
	"os"
)

func TestMain(m *testing.M) {
	Empower()
	os.Exit(m.Run())
}
`, `package gopwt

import (
	"testing"
	"os"
)

func TestMain(m *testing.M) {

	os.Exit(m.Run())
}
`)
}

func TestReplaceBinaryExpr(t *testing.T) {
	var parent ast.Expr
	var bin *ast.BinaryExpr

	parent = MustParseExpr("a(b+c)")
	bin = parent.(*ast.CallExpr).Args[0].(*ast.BinaryExpr)
	ReplaceBinaryExpr(parent, bin, MustParseExpr("x(1, 2)"))
	assert.OK(t, astToCode(parent) == "a(x(1, 2))")

	parent = MustParseExpr("struct{a int}{a: 1+2}").(*ast.CompositeLit).Elts[0]
	bin = parent.(*ast.KeyValueExpr).Value.(*ast.BinaryExpr)
	ReplaceBinaryExpr(parent, bin, MustParseExpr("x(1, 2)"))
	assert.OK(t, astToCode(parent) == "a: x(1, 2)")

	parent = MustParseExpr("a[1+2]")
	bin = parent.(*ast.IndexExpr).Index.(*ast.BinaryExpr)
	ReplaceBinaryExpr(parent, bin, MustParseExpr("x(1, 2)"))
	assert.OK(t, astToCode(parent) == "a[x(1, 2)]")

	parent = MustParseExpr("(1+2)")
	bin = parent.(*ast.ParenExpr).X.(*ast.BinaryExpr)
	ReplaceBinaryExpr(parent, bin, MustParseExpr("x(1, 2)"))
	assert.OK(t, astToCode(parent) == "(x(1, 2))")

	parent = MustParseExpr("1+1+2")
	bin = parent.(*ast.BinaryExpr).X.(*ast.BinaryExpr)
	ReplaceBinaryExpr(parent, bin, MustParseExpr("x(1, 2)"))
	assert.OK(t, astToCode(parent) == "x(1, 2) + 2")
}

func TestReplaceAllRawStringLitByStringLit(t *testing.T) {
	n := MustParseExpr(`func() string {
		return ` + "`" + `raw"string` + "`" + `
		}`)
	ReplaceAllRawStringLitByStringLit(n)
	assert.OK(t, astToCode(n) == `func() string {
	return "raw\"string"
}`)
}

func TestCreateRawStringLit(t *testing.T) {
	bq := "\"`\"" // back quote string literal "`"
	assert.OK(t, astToCode(CreateRawStringLit("foo")) == "`foo`")
	assert.OK(t, astToCode(CreateRawStringLit("foo`bar")) == strings.Join([]string{"`foo`", "`bar`"}, " + "+bq+" + "))
	assert.OK(t, astToCode(CreateRawStringLit("f`o`o")) == strings.Join([]string{"`f`", "`o`", "`o`"}, " + "+bq+" + "))
	assert.OK(t, astToCode(CreateRawStringLit("`ba`ba`ba`")) == strings.Join([]string{"``", "`ba`", "`ba`", "`ba`", "``"}, " + "+bq+" + "))
}

func TestCreateUntypedCallExprFromBuiltinCallExpr(t *testing.T) {
	var expr ast.Expr

	expr = CreateUntypedCallExprFromBuiltinCallExpr(ctx, MustParseExpr("make(typ)").(*ast.CallExpr))
	assert.OK(t, astToCode(expr) == "translatedassert.Bmake(translatedassert.RTOf(typ{}))")

	expr = CreateUntypedCallExprFromBuiltinCallExpr(ctx, MustParseExpr("new(typ)").(*ast.CallExpr))
	assert.OK(t, astToCode(expr) == "translatedassert.Bnew(translatedassert.RTOf(typ{}))")

	expr = CreateUntypedCallExprFromBuiltinCallExpr(ctx, MustParseExpr("cap(a)").(*ast.CallExpr))
	assert.OK(t, astToCode(expr) == "translatedassert.Bcap(a)")

	expr = CreateUntypedCallExprFromBuiltinCallExpr(ctx, MustParseExpr("complex(a, b)").(*ast.CallExpr))
	assert.OK(t, astToCode(expr) == "translatedassert.Bcomplex(a, b)")

	expr = CreateUntypedCallExprFromBuiltinCallExpr(ctx, MustParseExpr("copy(a, b)").(*ast.CallExpr))
	assert.OK(t, astToCode(expr) == "translatedassert.Bcopy(a, b)")

	expr = CreateUntypedCallExprFromBuiltinCallExpr(ctx, MustParseExpr("imag(a)").(*ast.CallExpr))
	assert.OK(t, astToCode(expr) == "translatedassert.Bimag(a)")

	expr = CreateUntypedCallExprFromBuiltinCallExpr(ctx, MustParseExpr("len(a)").(*ast.CallExpr))
	assert.OK(t, astToCode(expr) == "translatedassert.Blen(a)")

	expr = CreateUntypedCallExprFromBuiltinCallExpr(ctx, MustParseExpr("real(a)").(*ast.CallExpr))
	assert.OK(t, astToCode(expr) == "translatedassert.Breal(a)")
}

func TestCreateUntypedExprFromBinaryExpr(t *testing.T) {
	var f *ast.CallExpr

	f = CreateUntypedExprFromBinaryExpr(ctx, MustParseExpr("1 + 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "Op")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "`ADD`")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[2].(*ast.BasicLit).Value == "2")

	f = CreateUntypedExprFromBinaryExpr(ctx, MustParseExpr("1 - 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "Op")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "`SUB`")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[2].(*ast.BasicLit).Value == "2")

	f = CreateUntypedExprFromBinaryExpr(ctx, MustParseExpr("1 * 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "Op")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "`MUL`")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[2].(*ast.BasicLit).Value == "2")

	f = CreateUntypedExprFromBinaryExpr(ctx, MustParseExpr("1 / 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "Op")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "`QUO`")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[2].(*ast.BasicLit).Value == "2")

	f = CreateUntypedExprFromBinaryExpr(ctx, MustParseExpr("1 % 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "Op")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "`REM`")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[2].(*ast.BasicLit).Value == "2")

	f = CreateUntypedExprFromBinaryExpr(ctx, MustParseExpr("1 & 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "Op")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "`AND`")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[2].(*ast.BasicLit).Value == "2")

	f = CreateUntypedExprFromBinaryExpr(ctx, MustParseExpr("1 | 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "Op")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "`OR`")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[2].(*ast.BasicLit).Value == "2")

	f = CreateUntypedExprFromBinaryExpr(ctx, MustParseExpr("1 ^ 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "Op")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "`XOR`")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[2].(*ast.BasicLit).Value == "2")

	f = CreateUntypedExprFromBinaryExpr(ctx, MustParseExpr("1 &^ 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "Op")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "`ANDNOT`")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[2].(*ast.BasicLit).Value == "2")

	f = CreateUntypedExprFromBinaryExpr(ctx, MustParseExpr("1 << 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "OpShift")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "`SHL`")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[2].(*ast.BasicLit).Value == "2")

	f = CreateUntypedExprFromBinaryExpr(ctx, MustParseExpr("1 >> 2").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "OpShift")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "`SHR`")
	assert.OK(t, f.Args[1].(*ast.BasicLit).Value == "1")
	assert.OK(t, f.Args[2].(*ast.BasicLit).Value == "2")

	f = CreateUntypedExprFromBinaryExpr(ctx, MustParseExpr("true && true").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "Op")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "`LAND`")
	assert.OK(t, f.Args[1].(*ast.Ident).Name == "true")
	assert.OK(t, f.Args[2].(*ast.Ident).Name == "true")

	f = CreateUntypedExprFromBinaryExpr(ctx, MustParseExpr("false || false").(*ast.BinaryExpr)).(*ast.CallExpr)
	assert.OK(t, f.Fun.(*ast.SelectorExpr).Sel.Name == "Op")
	assert.OK(t, f.Args[0].(*ast.BasicLit).Value == "`LOR`")
	assert.OK(t, f.Args[1].(*ast.Ident).Name == "false")
	assert.OK(t, f.Args[2].(*ast.Ident).Name == "false")
}

func TestReplaceUnaryExpr(t *testing.T) {
	newExpr := MustParseExpr("rep()")

	{
		parent := MustParseExpr("a(!true)").(*ast.CallExpr)
		oldExpr := parent.Args[0].(*ast.UnaryExpr)
		ReplaceUnaryExpr(parent, oldExpr, newExpr)
		assert.OK(t, astToCode(parent) == "a(rep())")
	}
	{
		parent := MustParseExpr("map[int]bool{1:!true}").(*ast.CompositeLit).Elts[0].(*ast.KeyValueExpr)
		oldExpr := parent.Value.(*ast.UnaryExpr)
		ReplaceUnaryExpr(parent, oldExpr, newExpr)
		assert.OK(t, astToCode(parent) == "1: rep()")
	}
	{
		parent := MustParseExpr("a[!true]").(*ast.IndexExpr)
		oldExpr := parent.Index.(*ast.UnaryExpr)
		ReplaceUnaryExpr(parent, oldExpr, newExpr)
		assert.OK(t, astToCode(parent) == "a[rep()]")
	}
	{
		parent := MustParseExpr("(!true)").(*ast.ParenExpr)
		oldExpr := parent.X.(*ast.UnaryExpr)
		ReplaceUnaryExpr(parent, oldExpr, newExpr)
		assert.OK(t, astToCode(parent) == "(rep())")
	}
	{
		parent := MustParseExpr("!!true").(*ast.UnaryExpr)
		oldExpr := parent.X.(*ast.UnaryExpr)
		ReplaceUnaryExpr(parent, oldExpr, newExpr)
		assert.OK(t, astToCode(parent) == "!rep()")
	}
	{
		parent := MustParseExpr("!true + 1").(*ast.BinaryExpr)
		oldExpr := parent.X.(*ast.UnaryExpr)
		ReplaceUnaryExpr(parent, oldExpr, newExpr)
		assert.OK(t, astToCode(parent) == "rep() + 1")
	}
	{
		parent := MustParseExpr("1 + !true").(*ast.BinaryExpr)
		oldExpr := parent.Y.(*ast.UnaryExpr)
		ReplaceUnaryExpr(parent, oldExpr, newExpr)
		assert.OK(t, astToCode(parent) == "1 + rep()")
	}
}

func TestIsReflectDeepEqual(t *testing.T) {
	assert.OK(t, IsReflectDeepEqual(MustParseExpr("reflect.DeepEqual(1, 1)")) == true)
	assert.OK(t, IsReflectDeepEqual(MustParseExpr("reflect.DeepEqual2(1, 1)")) == false)
	assert.OK(t, IsReflectDeepEqual(MustParseExpr("reflect.DeepEqua(1, 1)")) == false)
}
