# gopwt

[![Drone Build Status](https://drone.io/github.com/ToQoz/gopwt/status.png)](https://drone.io/github.com/ToQoz/gopwt/latest)
[![Travis-CI Build Status](https://travis-ci.org/ToQoz/gopwt.svg?branch=master)](https://travis-ci.org/ToQoz/gopwt)
[![Coverage Status](https://img.shields.io/coveralls/ToQoz/gopwt.svg)](https://coveralls.io/r/ToQoz/gopwt?branch=master)

PowerAssert library for golang. This is out of goway(in my mind), but I'm going to put this on goway as possible as. Because I love it :)

## Usage

```
Usage of gopwt:
  -testdata="testdata": name of test data directories. e.g. -testdata testdata,migrations
  -v=false: This will be passed to `go test`
```

## Getting Started

### Install and Try

```
$ go get github.com/ToQoz/gopwt/...
$ cd your-go-project-path(in $GOPATH/src)
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
	"database/sql"
	"fmt"
	"github.com/ToQoz/gopwt/assert"
	"reflect"
	"testing"
)

func TestWithMessage(t *testing.T) {
	var receiver *struct{}
	receiver = nil
	assert.OK(t, receiver != nil, "receiver should not be nil")
}

func TestBasicLit(t *testing.T) {
	assert.OK(t, "a" == "b")
	assert.OK(t, 1 == 2)

	a := 1
	b := 2
	c := 3
	assert.OK(t, a+c == b)
	assert.OK(t, (a+c)+a == b)
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
	assert.OK(t, []int{
		1,
		2,
	}[index] == 3)
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

func TestCallWithNonIdempotentFunc(t *testing.T) {
	i := 0
	incl := func() int {
		i++
		return i
	}

	assert.OK(t, incl()+incl() == incl()+incl())
	assert.OK(t, incl() == incl())
	assert.OK(t, incl() == incl())
	assert.OK(t, incl() == incl())
	assert.OK(t, incl() == incl())
	assert.OK(t, (incl() == incl()) != (incl() == incl()))

	i2 := 0
	incl2 := func(i3 int) int {
		i2 += i3
		return i2
	}

	assert.OK(t, incl2(incl2(2)) == 10)
}

func TestPkgValue(t *testing.T) {
	assert.OK(t, sql.ErrNoRows == fmt.Errorf("error"))
}
```

```
$ gopwt
--- FAIL: TestWithMessage (0.00s)
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:14
		assert.OK(t, receiver != nil, "receiver should not be nil")
		             |        |  |
		             |        |  <nil>
		             |        false
		             (*struct {})(nil)
		
		Assersion messages:
			- receiver should not be nil
--- FAIL: TestBasicLit (0.00s)
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:17
		assert.OK(t, "a" == "b")
		                 |
		                 false
		
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:18
		assert.OK(t, 1 == 2)
		               |
		               false
		
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:22
		assert.OK(t, a+c == b)
		             ||| |  |
		             ||| |  2
		             ||| false
		             ||3
		             |4
		             1
		
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:23
		assert.OK(t, (a+c)+a == b)
		              ||| || |  |
		              ||| || |  2
		              ||| || false
		              ||| |1
		              ||| 5
		              ||3
		              |4
		              1
		
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:24
		assert.OK(t, "foo\nbar" == "bar")
		             |          |
		             |          false
		             "foo
		             bar"
		
--- FAIL: TestMapType (0.00s)
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:29
		assert.OK(t, reflect.DeepEqual(map[string]string{}, map[string]string{"a": "a", k: v}))
		             |                                                                  |  |
		             |                                                                  |  "b------value"
		             |                                                                  "b--------key"
		             false
		
--- FAIL: TestArrayType (0.00s)
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:33
		assert.OK(t, []int{1, 2}[index] == 3)
		                         |      |
		                         |      false
		                         1
		
--- FAIL: TestStructType (0.00s)
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:37
		assert.OK(t, reflect.DeepEqual(struct{ Name string }{foox}, struct{ Name string }{"foo"}))
		             |                                       |
		             |                                       "foo------x"
		             false
		
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:38
		assert.OK(t, reflect.DeepEqual(struct{ Name string }{Name: foox}, struct{ Name string }{Name: "foo"}))
		             |                                             |
		             |                                             "foo------x"
		             false
		
--- FAIL: TestNestedCallExpr (0.00s)
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:44
		assert.OK(t, rev(rev(rev(true))))
		             |   |   |   |
		             |   |   |   true
		             |   |   false
		             |   true
		             false
		
--- FAIL: TestCallWithNonIdempotentFunc (0.00s)
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:52
		assert.OK(t, incl()+incl() == incl()+incl())
		             |     ||      |  |     ||
		             |     ||      |  |     |4
		             |     ||      |  |     7
		             |     ||      |  3
		             |     ||      false
		             |     |2
		             |     3
		             1
		
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:53
		assert.OK(t, incl() == incl())
		             |      |  |
		             |      |  6
		             |      false
		             5
		
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:54
		assert.OK(t, incl() == incl())
		             |      |  |
		             |      |  8
		             |      false
		             7
		
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:55
		assert.OK(t, incl() == incl())
		             |      |  |
		             |      |  10
		             |      false
		             9
		
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:56
		assert.OK(t, incl() == incl())
		             |      |  |
		             |      |  12
		             |      false
		             11
		
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:57
		assert.OK(t, (incl() == incl()) != (incl() == incl()))
		              |      |  |       |   |      |  |
		              |      |  |       |   |      |  16
		              |      |  |       |   |      false
		              |      |  |       |   15
		              |      |  |       false
		              |      |  14
		              |      false
		              13
		
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:63
		assert.OK(t, incl2(incl2(2)) == 10)
		             |     |         |
		             |     |         false
		             |     2
		             4
		
--- FAIL: TestPkgValue (0.00s)
	assert.go:61: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:66
		assert.OK(t, sql.ErrNoRows == fmt.Errorf("error"))
		                 |         |  |
		                 |         |  &errors.errorString{s:"error"}
		                 |         false
		                 &errors.errorString{s:"sql: no rows in result set"}
		
FAIL
FAIL	github.com/ToQoz/gopwt/_example	0.008s
exit status 1
```

## See Also

- [Why does Go not have assertions?](http://golang.org/doc/faq#assertions)
- http://dontmindthelanguage.wordpress.com/2009/12/11/groovy-1-7-power-assert/
- https://github.com/twada/power-assert
