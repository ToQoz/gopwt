# gopwt

## Getting Started

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

## Attention

Calling non-idempotent func in `assert.OK()` make a difference between real value and output value.(I'm trying to fix)

For example.

```go

package main

import (
	"github.com/ToQoz/gopwt/assert"
	"testing"
)

func TestBad(t *testing.T) {
	i := 0
	incl := func() int {
		i++
		return i
	}

	assert.OK(t, incl() == incl())
}
```

```

--- FAIL: TestBad (0.00 seconds)
	assert.go:49: [FAIL] /Users/toqoz/_go/src/github.com/ToQoz/gopwt/_example/badexample_test.go:15
		assert.OK(t, incl() == incl())
		             |      |  |
		             |      |  6
		             |      false
		             3
```

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
