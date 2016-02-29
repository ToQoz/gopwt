package main

import (
	"github.com/ToQoz/gopwt/assert"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"strings"
	"testing"
)

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
	types, err := getTypeInfo("./tdata/is_type_conversion_test", "github.com/ToQoz/gopwt/tdata/is_type_conversion_test", strings.Split(os.Getenv("GOPATH"), ":")[0]+"/src", fset, []*ast.File{f})
	assert.Require(t, err == nil)

	// fmt.Println(string([]byte(hello())))
	expr := f.Decls[1].(*ast.FuncDecl).Body.List[0].(*ast.ExprStmt).X

	fmtPrintln := expr.(*ast.CallExpr)
	assert.OK(t, isTypeConversion(types, fmtPrintln) == false, "fmt.Println(x) is NOT type conversion")
	stringConv := fmtPrintln.Args[0].(*ast.CallExpr)
	assert.OK(t, isTypeConversion(types, stringConv) == true, "string(x) is type conversion")
	bytesConv := stringConv.Args[0].(*ast.CallExpr)
	assert.OK(t, isTypeConversion(types, bytesConv) == true, "[]byte(x) is type conversion")

	// fmt.Println(http.Handler(nil))
	expr = f.Decls[1].(*ast.FuncDecl).Body.List[1].(*ast.ExprStmt).X
	httpHandlerConv := expr.(*ast.CallExpr).Args[0].(*ast.CallExpr)
	assert.OK(t, isTypeConversion(types, httpHandlerConv) == true, "http.Handler(x) is type conversion")
}
