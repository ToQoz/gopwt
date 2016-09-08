package internal_test

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"testing"

	"github.com/ToQoz/gopwt/assert"
	. "github.com/ToQoz/gopwt/translator/internal"
)

func TestDontPanic_OnTypeConversion(t *testing.T) {
	a := func() rune {
		return 'a'
	}
	assert.OK(t, string(a()) == string(a()))

	c := 0
	incl := func() int {
		c++
		return c
	}

	assert.OK(t, incl() == incl()-1)
	assert.OK(t, int32(incl()) == int32(incl()-1))
	assert.OK(t, int32(incl()) == int32(incl()-1))
}

func TestCopyFile(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	CopyFile("./testdata/rewrite_file_tests/simple.go", buf)

	assert.OK(t, buf.String() == `package main

import (
	"testing"

	"github.com/ToQoz/gopwt/assert"
)

func TestSimple(t *testing.T) {
	func() {
		assert.OK(t, 1 == 1, "1 is 1")
	}()
}
`)
}

func TestRewriteFile(t *testing.T) {
	var file string

	file = "./testdata/rewrite_file_tests/simple.go"
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, 0)
	assert.Require(t, err == nil)

	expected, err := readFileString(file + ".tr.txt")
	assert.Require(t, err == nil)

	buf := bytes.NewBuffer([]byte{})
	RewriteFile(nil, fset, fset, f, f, buf)
	got := buf.String()

	assert.OK(t, got == expected)
}

func TestCreateReflectTypeExprFromTypeExpr(t *testing.T) {
	// built in type
	assert.OK(t, "translatedassert.RVOf(new(string)).Elem().Type()" == astToCode(CreateReflectTypeExprFromTypeExpr(MustParseExpr("string"))))
	assert.OK(t, "translatedassert.RVOf(new(int)).Elem().Type()" == astToCode(CreateReflectTypeExprFromTypeExpr(MustParseExpr("int"))))
	// chan
	assert.OK(t, "translatedassert.RVOf(new(chan int)).Elem().Type()" == astToCode(CreateReflectTypeExprFromTypeExpr(MustParseExpr("chan int"))))
	// map, slice
	assert.OK(t, "translatedassert.RTOf([]string{})" == astToCode(CreateReflectTypeExprFromTypeExpr(MustParseExpr("[]string"))))
	assert.OK(t, "translatedassert.RTOf(map[string]string{})" == astToCode(CreateReflectTypeExprFromTypeExpr(MustParseExpr("map[string]string"))))
	// func
	assert.OK(t, "translatedassert.RTOf(func(){})" == astToCode(CreateReflectTypeExprFromTypeExpr(MustParseExpr("func()"))))
	// other type
	assert.OK(t, "translatedassert.RTOf(foo{})" == astToCode(CreateReflectTypeExprFromTypeExpr(MustParseExpr("foo"))))
}

func TestExtractPrintExprs_SingleLineStringLit(t *testing.T) {
	ps := ExtractPrintExprs(nil, "", 0, 0, nil, MustParseExpr(`"foo" == "bar"`))
	assert.OK(t, len(ps) == 3)
	// ==
	assert.OK(t, ps[1].Pos == len(`"foo" `)+1)
}

func TestExtractPrintExprs_MultiLineStringLit(t *testing.T) {
	ps := ExtractPrintExprs(nil, "", 0, 0, nil, MustParseExpr(`"foo\nbar" == "bar"`))
	assert.OK(t, len(ps) == 3)
	assert.OK(t, ps[0].Pos == 1)
	assert.OK(t, ps[0].Expr.(*ast.BasicLit).Value == `"foo\nbar"`)
	assert.OK(t, ps[2].Pos == len(`"foo\nbar" == `)+1)
	assert.OK(t, ps[2].Expr.(*ast.BasicLit).Value == `"bar"`)
}

func TestExtractPrintExprs_UnaryExpr(t *testing.T) {
	// !a -> !translatedassert.RVBool(translatedassert.RVOf(a))
	ps := ExtractPrintExprs(nil, "", 0, 0, nil, MustParseExpr("!a"))
	assert.OK(t, len(ps) == 2)
	assert.OK(t, ps[0].Pos == 1)
	assert.OK(t, ps[0].Expr.(*ast.UnaryExpr).X.(*ast.CallExpr).Fun.(*ast.SelectorExpr).X.(*ast.Ident).Name == "translatedassert")
	assert.OK(t, ps[0].Expr.(*ast.UnaryExpr).X.(*ast.CallExpr).Fun.(*ast.SelectorExpr).Sel.Name == "RVBool")
	assert.OK(t, ps[1].Pos == 2)
	assert.OK(t, ps[1].Expr.(*ast.Ident).Name == "a")
}

func TestExtractPrintExprs_StarExpr(t *testing.T) {
	ps := ExtractPrintExprs(nil, "", 0, 0, nil, MustParseExpr("*a"))
	assert.OK(t, len(ps) == 2)
	assert.OK(t, ps[0].Pos == 1)
	assert.OK(t, ps[0].Expr.(*ast.StarExpr).X.(*ast.Ident).Name == "a")
	assert.OK(t, ps[1].Pos == 2)
	assert.OK(t, ps[1].Expr.(*ast.Ident).Name == "a")
}

func TestExtractPrintExprs_SliceExpr(t *testing.T) {
	ps := ExtractPrintExprs(nil, "", 0, 0, nil, MustParseExpr(`"foo"[a1:a2]`))
	assert.OK(t, len(ps) == 2)
	assert.OK(t, ps[0].Pos == len(`"foo"[`)+1)
	assert.OK(t, ps[0].Expr.(*ast.Ident).Name == "a1")
	assert.OK(t, ps[1].Pos == len(`"foo"[a1:`)+1)
	assert.OK(t, ps[1].Expr.(*ast.Ident).Name == "a2")

	ps = ExtractPrintExprs(nil, "", 0, 0, nil, MustParseExpr(`"foo"[a1:a2:a3]`))
	assert.OK(t, len(ps) == 3)
	assert.OK(t, ps[0].Pos == len(`"foo"[`)+1)
	assert.OK(t, ps[0].Expr.(*ast.Ident).Name == "a1")
	assert.OK(t, ps[1].Pos == len(`"foo"[a1:`)+1)
	assert.OK(t, ps[1].Expr.(*ast.Ident).Name == "a2")
	assert.OK(t, ps[2].Pos == len(`"foo"[a1:a2:`)+1)
	assert.OK(t, ps[2].Expr.(*ast.Ident).Name == "a3")
}

func TestExtractPrintExprs_IndexExpr(t *testing.T) {
	ps := ExtractPrintExprs(nil, "", 0, 0, nil, MustParseExpr("ary[i] == ary2[i2]"))
	assert.OK(t, len(ps) == 7)

	// ary
	assert.OK(t, ps[0].Pos == 1)
	assert.OK(t, ps[0].Expr.(*ast.Ident).Name == "ary")
	// ary[i]
	assert.OK(t, ps[1].Pos == len("ary")+1)
	assert.OK(t, ps[1].Expr.(*ast.IndexExpr).X == ps[0].Expr)
	assert.OK(t, ps[1].Expr.(*ast.IndexExpr).Index == ps[2].Expr)
	// i
	assert.OK(t, ps[2].Pos == len("ary[")+1)
	assert.OK(t, ps[2].Expr.(*ast.Ident).Name == "i")

	// ==
	assert.OK(t, ps[3].Pos == len("ary[i] ")+1)

	// ary2
	assert.OK(t, ps[4].Pos == len("ary[i] == ")+1)
	assert.OK(t, ps[4].Expr.(*ast.Ident).Name == "ary2")
	// ary2[i2]
	assert.OK(t, ps[5].Expr.(*ast.IndexExpr).X == ps[4].Expr)
	assert.OK(t, ps[5].Expr.(*ast.IndexExpr).Index == ps[6].Expr)
	// i2
	assert.OK(t, ps[6].Pos == len("ary[i] == ary2[")+1)
	assert.OK(t, ps[6].Expr.(*ast.Ident).Name == "i2")
}

func TestExtractPrintExprs_ArrayType(t *testing.T) {
	ps := ExtractPrintExprs(nil, "", 0, 0, nil, MustParseExpr("reflect.DeepEqual([]string{c}, []string{})"))
	assert.OK(t, len(ps) == 4)
	// reflect.DeepEqual
	assert.OK(t, ps[0].Pos == 1)
	// []string{c}
	assert.OK(t, ps[1].Pos == len("reflect.DeepEqual(")+1)
	// c
	assert.OK(t, ps[2].Pos == len("reflect.DeepEqual([]string{")+1)
	assert.OK(t, ps[2].Expr.(*ast.Ident).Name == "c")
	// []string{}
	assert.OK(t, ps[3].Pos == len("reflect.DeepEqual([]string{c}, ")+1)

	ps = ExtractPrintExprs(nil, "", 0, 0, nil, MustParseExpr("reflect.DeepEqual([4]string{d}, []string{})"))
	assert.OK(t, len(ps) == 4)
	// reflect.DeeepEqual
	assert.OK(t, ps[0].Pos == 1)
	// [4]string{d}
	assert.OK(t, ps[1].Pos == len("reflect.DeepEqual(")+1)
	// d
	assert.OK(t, ps[2].Pos == len("reflect.DeepEqual([4]string{")+1)
	// []string{}
	assert.OK(t, ps[3].Pos == len("reflect.DeepEqual([4]string{d}, ")+1)
}

func TestExtractPrintExprs_MapType(t *testing.T) {
	ps := ExtractPrintExprs(nil, "", 0, 0, nil, MustParseExpr("reflect.DeepEqual(map[string]string{a:b}, map[string]string{})"))
	assert.OK(t, len(ps) == 5)

	// reflect.DeepEqual
	assert.OK(t, ps[0].Pos == 1)

	// map[string]string{a:b}
	assert.OK(t, ps[1].Pos == len("reflect.DeepEqual(")+1)

	// a
	assert.OK(t, ps[2].Pos == len("reflect.DeepEqual(map[string]string{")+1)
	assert.OK(t, ps[2].Expr.(*ast.Ident).Name == "a")

	// b
	assert.OK(t, ps[3].Pos == len("reflect.DeepEqual(map[string]string{a:")+1)
	assert.OK(t, ps[3].Expr.(*ast.Ident).Name == "b")

	// map[string]string{}
	assert.OK(t, ps[4].Pos == len("reflect.DeepEqual(map[string]string{a:b}, ")+1)
}

func TestExtractPrintExprs_StructType(t *testing.T) {
	ps := ExtractPrintExprs(nil, "", 0, 0, nil, MustParseExpr("reflect.DeepEqual(struct{Name string}{}, struct{Name string}{Name: foo})"))
	assert.OK(t, len(ps) == 4)
	// reflect.DeepEqual
	assert.OK(t, ps[0].Pos == 1)
	// struct{Name string}{}
	assert.OK(t, ps[1].Pos == len("reflect.DeepEqual(")+1)
	// struct{Name string}{Name: foo}
	assert.OK(t, ps[2].Pos == len("reflect.DeepEqual(struct{Name string}{}, ")+1)
	// foo
	assert.OK(t, ps[3].Pos == len("reflect.DeepEqual(struct{Name string}{}, struct{Name string}{Name: ")+1)
	assert.OK(t, ps[3].Expr.(*ast.Ident).Name == "foo")
}

func TestConvertFuncCallToMemorized(t *testing.T) {
	expected := `translatedassert.FRVInterface(translatedassert.MFCall("", 0, 1, translatedassert.RVOf(f), translatedassert.RVOf(a), translatedassert.RVOf(b)))`
	assert.OK(t, astToCode(CreateMemorizedFuncCall("", 0, MustParseExpr("f(a, b)").(*ast.CallExpr), "Interface")) == expected)

	expected = `translatedassert.FRVBool(translatedassert.MFCall("", 0, 1, translatedassert.RVOf(f), translatedassert.RVOf(b)))`
	assert.OK(t, astToCode(CreateMemorizedFuncCall("", 0, MustParseExpr("f(b)").(*ast.CallExpr), "Bool")) == expected)
}

func Test_CreateUntypedExprFromBinaryExpr_and_ReplaceBinaryExpr(t *testing.T) {
	// CallExpr
	func() {
		parent := MustParseExpr("f(b + a)").(*ast.CallExpr)
		newExpr := CreateUntypedExprFromBinaryExpr(parent.Args[0].(*ast.BinaryExpr))
		if newExpr != parent.Args[0].(*ast.BinaryExpr) {
			ReplaceBinaryExpr(parent, parent.Args[0].(*ast.BinaryExpr), newExpr)
		}
		assert.OK(t, astToCode(parent) == `f(translatedassert.OpADD(b, a))`)
		assert.OK(t, astToCode(newExpr) == `translatedassert.OpADD(b, a)`)
	}()
	func() {
		parent := MustParseExpr("f(b, b + a)").(*ast.CallExpr)
		newExpr := CreateUntypedExprFromBinaryExpr(parent.Args[1].(*ast.BinaryExpr))
		if newExpr != parent.Args[1].(*ast.BinaryExpr) {
			ReplaceBinaryExpr(parent, parent.Args[1].(*ast.BinaryExpr), newExpr)
		}
		assert.OK(t, astToCode(parent) == `f(b, translatedassert.OpADD(b, a))`)
		assert.OK(t, astToCode(newExpr) == `translatedassert.OpADD(b, a)`)
	}()
	// ParentExpr
	func() {
		parent := MustParseExpr("(b + a)").(*ast.ParenExpr)
		newExpr := CreateUntypedExprFromBinaryExpr(parent.X.(*ast.BinaryExpr))
		if newExpr != parent.X.(*ast.BinaryExpr) {
			ReplaceBinaryExpr(parent, parent.X.(*ast.BinaryExpr), newExpr)
		}
		assert.OK(t, astToCode(parent) == `(translatedassert.OpADD(b, a))`)
		assert.OK(t, astToCode(newExpr) == `translatedassert.OpADD(b, a)`)
	}()
	// BinaryExpr
	func() {
		parent := MustParseExpr("b + a == c + d").(*ast.BinaryExpr)
		newExpr := CreateUntypedExprFromBinaryExpr(parent.X.(*ast.BinaryExpr))
		if newExpr != parent.X.(*ast.BinaryExpr) {
			ReplaceBinaryExpr(parent, parent.X.(*ast.BinaryExpr), newExpr)
		}
		assert.OK(t, astToCode(parent) == `translatedassert.OpADD(b, a) == c+d`)
		assert.OK(t, astToCode(newExpr) == `translatedassert.OpADD(b, a)`)
		newExpr = CreateUntypedExprFromBinaryExpr(parent.Y.(*ast.BinaryExpr))
		if newExpr != parent.Y.(*ast.BinaryExpr) {
			ReplaceBinaryExpr(parent, parent.Y.(*ast.BinaryExpr), newExpr)
		}
		assert.OK(t, astToCode(parent) == `translatedassert.OpADD(b, a) == translatedassert.OpADD(c, d)`)
		assert.OK(t, astToCode(newExpr) == `translatedassert.OpADD(c, d)`)
	}()
	// KeyValuePair
	func() {
		_parent := MustParseExpr("map[string]string{a + b: c + d}").(*ast.CompositeLit)
		parent := _parent.Elts[0].(*ast.KeyValueExpr)
		newExpr := CreateUntypedExprFromBinaryExpr(parent.Key.(*ast.BinaryExpr))
		if newExpr != parent.Key.(*ast.BinaryExpr) {
			ReplaceBinaryExpr(parent, parent.Key.(*ast.BinaryExpr), newExpr)
		}
		assert.OK(t, astToCode(parent) == `translatedassert.OpADD(a, b): c + d`)
		assert.OK(t, astToCode(newExpr) == `translatedassert.OpADD(a, b)`)
		newExpr = CreateUntypedExprFromBinaryExpr(parent.Value.(*ast.BinaryExpr))
		if newExpr != parent.Value.(*ast.BinaryExpr) {
			ReplaceBinaryExpr(parent, parent.Value.(*ast.BinaryExpr), newExpr)
		}
		assert.OK(t, astToCode(parent) == `translatedassert.OpADD(a, b): translatedassert.OpADD(c, d)`)
		assert.OK(t, astToCode(newExpr) == `translatedassert.OpADD(c, d)`)
	}()
	// IndexExpr
	func() {
		parent := MustParseExpr("a[a+b]").(*ast.IndexExpr)
		newExpr := CreateUntypedExprFromBinaryExpr(parent.Index.(*ast.BinaryExpr))
		if newExpr != parent.Index.(*ast.BinaryExpr) {
			ReplaceBinaryExpr(parent, parent.Index.(*ast.BinaryExpr), newExpr)
		}
		assert.OK(t, astToCode(parent) == `a[translatedassert.OpADD(a, b)]`)
		assert.OK(t, astToCode(newExpr) == `translatedassert.OpADD(a, b)`)
	}()
}

func TestResultPosOf(t *testing.T) {
	assert.OK(t, int(ResultPosOf(MustParseExpr("a[2]"))) == len("a["))
	assert.OK(t, int(ResultPosOf(MustParseExpr("x.Println"))) == len("x.P"))
	assert.OK(t, int(ResultPosOf(MustParseExpr("1 == 2"))) == len("1 ="))
	assert.OK(t, int(ResultPosOf(MustParseExpr("(foo + bar)"))) == len("(foo +"))
}

func astToCode(a ast.Node) string {
	b := []byte{}
	buf := bytes.NewBuffer(b)
	printer.Fprint(buf, token.NewFileSet(), a)
	return buf.String()
}

func MustParseExpr(s string) ast.Expr {
	e, err := parser.ParseExpr(s)
	if err != nil {
		panic(err)
	}

	return e
}

func readFileString(file string) (string, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
