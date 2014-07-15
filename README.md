# gopwt

PowerAssert library for golang. This is out of goway(in my mind), but I'm going to put this on goway as possible as. Because I love it :)

## Getting Started

### Install and Try

```
$ go get github.com/ToQoz/gopwt/...
$ cd your-go-project-path
$ vi main_test.go
$ cat main_test.go
package main

import (
	"github.com/ToQoz/gopwt/assert"
	"testing"
)

func TestFoo(t *testing.T) {
	a := "a"
	b := "b"
	assert.OK(t, a == b)
}
$ gopwt
```

### Update

`go get -u github.com/ToQoz/gopwt/...`

## Example

```go
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
```

```
$ gopwt
--- FAIL: TestBasicLit (0.00 seconds)
	assert.go:49: [FAIL] /Users/toqoz/_go/src/github.com/ToQoz/gopwt/_example/example_test.go:10
		assert.OK(t, "a" == "b")
		                 |
		                 false

	assert.go:49: [FAIL] /Users/toqoz/_go/src/github.com/ToQoz/gopwt/_example/example_test.go:11
		assert.OK(t, 1 == 2)
		               |
		               false

	assert.go:49: [FAIL] /Users/toqoz/_go/src/github.com/ToQoz/gopwt/_example/example_test.go:16
		assert.OK(t, a+c == b)
		             ||| |  |
		             ||| |  2
		             ||| false
		             ||3
		             |4
		             1

	assert.go:49: [FAIL] /Users/toqoz/_go/src/github.com/ToQoz/gopwt/_example/example_test.go:17
		assert.OK(t, "foo\nbar" == "bar")
		             |          |
		             |          false
		             "foo
		             bar"

--- FAIL: TestMapType (0.00 seconds)
	assert.go:49: [FAIL] /Users/toqoz/_go/src/github.com/ToQoz/gopwt/_example/example_test.go:24
		assert.OK(t, reflect.DeepEqual(map[string]string{}, map[string]string{"a": "a", k: v}))
		             |                                                                  |  |
		             |                                                                  |  "b------value"
		             |                                                                  "b--------key"
		             false

--- FAIL: TestArrayType (0.00 seconds)
	assert.go:49: [FAIL] /Users/toqoz/_go/src/github.com/ToQoz/gopwt/_example/example_test.go:32
		assert.OK(t, []int{1, 2}[index] == 3)
		                         |      |
		                         |      false
		                         1

--- FAIL: TestStructType (0.00 seconds)
	assert.go:49: [FAIL] /Users/toqoz/_go/src/github.com/ToQoz/gopwt/_example/example_test.go:37
		assert.OK(t, reflect.DeepEqual(struct{ Name string }{foox}, struct{ Name string }{"foo"}))
		             |                                       |
		             |                                       "foo------x"
		             false

	assert.go:49: [FAIL] /Users/toqoz/_go/src/github.com/ToQoz/gopwt/_example/example_test.go:38
		assert.OK(t, reflect.DeepEqual(struct{ Name string }{Name: foox}, struct{ Name string }{Name: "foo"}))
		             |                                             |
		             |                                             "foo------x"
		             false

FAIL
FAIL	github.com/ToQoz/gopwt/_example	0.012s
exit status 1
```

## See Also

- [Why does Go not have assertions?](http://golang.org/doc/faq#assertions)
- http://dontmindthelanguage.wordpress.com/2009/12/11/groovy-1-7-power-assert/
- https://github.com/twada/power-assert
