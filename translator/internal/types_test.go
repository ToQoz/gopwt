package internal_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
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

	err := filepath.Walk("testdata", func(path string, fInfo os.FileInfo, err error) error {
		if fInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil
		}

		rel, err := filepath.Rel("testdata", path)
		if err != nil {
			return err
		}
		outPath := filepath.Join("tdata", rel)

		if fInfo.IsDir() {
			di, err := os.Stat(path)
			if err != nil {
				return err
			}
			err = os.Mkdir(outPath, di.Mode())
			if err != nil {
				return err
			}
			return nil
		}

		in, err := os.OpenFile(path, os.O_RDWR, fInfo.Mode())
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(outPath, os.O_RDWR|os.O_CREATE, fInfo.Mode())
		if err != nil {
			return err
		}
		defer out.Close()

		io.Copy(out, in)
		return nil
	})

	assert.Require(t, err == nil)
	defer os.RemoveAll("tdata")

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filepath.Join("tdata", "is_type_conversion_test", "main.go"), nil, 0)
	assert.Require(t, err == nil)
	types, err := GetTypeInfo("", false, filepath.Join("tdata", "is_type_conversion_test"), "github.com/ToQoz/gopwt/translator/internal/tdata/is_type_conversion_test", strings.Split(os.Getenv("GOPATH"), string(filepath.ListSeparator))[0]+"/src", fset, []*ast.File{f})
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
