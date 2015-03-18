package main

import (
	"go/ast"
	"go/parser"
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
