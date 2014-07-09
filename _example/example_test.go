package main

import (
	"github.com/ToQoz/gopwt/assert"
	"reflect"
	"testing"
)

func TestBasicLit(t *testing.T) {
	assert.OK(t, "a" == "b")
	assert.OK(t, 1 == 2)

	a := 1
	b := 2
	c := 3
	assert.OK(t, a+c == b)
	assert.OK(t, `foo
bar` == "bar")
}

func TestMapType(t *testing.T) {
	k := "b--------key"
	v := "b------value"
	assert.OK(t, reflect.DeepEqual(map[string]string{}, map[string]string{
		"a": "a",
		k:   v,
	}))
}

func TestArrayType(t *testing.T) {
	index := 1
	assert.OK(t, []int{1, 2}[index] == 3)
}

func TestStructType(t *testing.T) {
	foox := "foo------x"
	assert.OK(t, reflect.DeepEqual(struct{ Name string }{foox}, struct{ Name string }{"foo"}))
	assert.OK(t, reflect.DeepEqual(struct{ Name string }{Name: foox}, struct{ Name string }{Name: "foo"}))
}

func TestNestedCallExpr(t *testing.T) {
	rev := func(a bool) bool {
		return !a
	}

	assert.OK(t, rev(rev(rev(true))))
}
