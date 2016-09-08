package internal_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/ToQoz/gopwt/assert"
	. "github.com/ToQoz/gopwt/translator/internal"
)

func TestDeterminantExprOfIsTypeConversion(t *testing.T) {
	var d ast.Expr

	d = DeterminantExprOfIsTypeConversion(MustParseExpr("(string(x))"))
	assert.OK(t, astToCode(d) == "string")

	d = DeterminantExprOfIsTypeConversion(MustParseExpr("*string(x)"))
	assert.OK(t, astToCode(d) == "string")

	d = DeterminantExprOfIsTypeConversion(MustParseExpr("string(x)"))
	assert.OK(t, astToCode(d) == "string")

	d = DeterminantExprOfIsTypeConversion(MustParseExpr("http.Handler(x)"))
	assert.OK(t, astToCode(d) == "Handler")
}

func TestIsTypeConversion(t *testing.T) {
	// go install fails under ./testdata
	//   (go install: no install location for directory github.com/ToQoz/gopwt/testdata/is_type_conversion_test outside GOPATH)
	// So copy to ./tdata temporary
	cp := exec.Command("cp", "-r", "./testdata", "./tdata")
	cp.Stdout = os.Stdout
	cp.Stderr = os.Stderr
	err := cp.Run()
	assert.Require(t, err == nil)
	defer os.RemoveAll("./tdata")

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "./tdata/is_type_conversion_test/main.go", nil, 0)
	assert.Require(t, err == nil)
	types, err := GetTypeInfo("./tdata/is_type_conversion_test", "github.com/ToQoz/gopwt/translator/internal/tdata/is_type_conversion_test", strings.Split(os.Getenv("GOPATH"), ":")[0]+"/src", fset, []*ast.File{f})
	assert.Require(t, err == nil)

	// fmt.Println(string([]byte(hello())))
	expr := f.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.ExprStmt).X

	fmtPrintln := expr.(*ast.CallExpr)
	assert.OK(t, IsTypeConversion(types, fmtPrintln) == false, "fmt.Println(x) is NOT type conversion")
	stringConv := fmtPrintln.Args[0].(*ast.CallExpr)
	assert.OK(t, IsTypeConversion(types, stringConv) == true, "string(x) is type conversion")
	bytesConv := stringConv.Args[0].(*ast.CallExpr)
	assert.OK(t, IsTypeConversion(types, bytesConv) == true, "[]byte(x) is type conversion")

	// fmt.Println(http.Handler(nil))
	expr = f.Decls[1].(*ast.FuncDecl).Body.List[1].(*ast.ExprStmt).X
	httpHandlerConv := expr.(*ast.CallExpr).Args[0].(*ast.CallExpr)
	assert.OK(t, IsTypeConversion(types, httpHandlerConv) == true, "http.Handler(x) is type conversion")
}
