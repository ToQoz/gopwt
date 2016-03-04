# gopwt

[![Drone Build Status](https://drone.io/github.com/ToQoz/gopwt/status.png)](https://drone.io/github.com/ToQoz/gopwt/latest)
[![Travis-CI Build Status](https://travis-ci.org/ToQoz/gopwt.svg?branch=master)](https://travis-ci.org/ToQoz/gopwt)

|package|coverage|
|-------|-----|
|gopwt v0.2.0| [![](https://gocover.io/_badge/github.com/toqoz/gopwt?v0.2.0)](https://gocover.io/github.com/toqoz/gopwt)|
|gopwt/assert v0.2.0| [![](https://gocover.io/_badge/github.com/toqoz/gopwt/assert?v0.2.0)](https://gocover.io/github.com/toqoz/gopwt/assert)|
|gopwt/translatedassert v0.2.0| [![](https://gocover.io/_badge/github.com/toqoz/gopwt/translatedassert?v0.2.0)](https://gocover.io/github.com/toqoz/gopwt/translatedassert)|

PowerAssert library for golang. This is out of goway(in my mind), but I'm going to put this on goway as possible as. Because I love it :)

![logo](http://toqoz.net/art/images/gopwt.svg)

## Supported go versions

See [.travis.yml](/.travis.yml)

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
$ mkdir $GOPATH/src/$(whoami)/gopwtexample
$ cd $GOPATH/src/$(whoami)/gopwtexample
$ cat <<EOF > main_test.go
package main

import (
	"github.com/ToQoz/gopwt/assert"
	"testing"
)

func TestFoo(t *testing.T) {
	a := "a"
	b := "b"
	assert.OK(t, a == b, "a should equal to b")
}
EOF
$ gopwt
--- FAIL: TestFoo (0.00s)
	assert.go:61: FAIL /var/folders/f8/0gm3xlgn1q12_zt7kxmzfj480000gn/T/109993308/src/github.com/ToQoz/gopwtexample/main_test.go:11
		assert.OK(t, a == b, "a should equal to b")
		             | |  |
		             | |  "b"
		             | false
		             "a"

		Assersion messages:
			- a should equal to b

		--- [string] expected
		--- [string] got
		@@ -1,1 +1,1@@
		-b
		+a


FAIL
FAIL    github.com/ToQoz/gopwtexample        0.008s
exit status 1
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

func TestStringDiff(t *testing.T) {
	assert.OK(t, "supersoper" == "supersuper")

	expected := `foo
baz
bar2
bar`
	got := `foo
baz
bar
bar`
	assert.OK(t, got == expected)

	expected = `<div>
<div>
foo
</div>
</div>`
	got = `<div>
bar
</div>`
	assert.OK(t, got == expected)
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
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:14
		assert.OK(t, receiver != nil, "receiver should not be nil")
		             |        |  |
		             |        |  <nil>
		             |        false
		             (*struct {})(nil)
		
		Assersion messages:
			- receiver should not be nil
		
--- FAIL: TestBasicLit (0.00s)
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:17
		assert.OK(t, "a" == "b")
		                 |
		                 false
		
		--- [string] expected
		+++ [string] got
		@@ -1,1 +1,1@@
		-b
		+a
		
		
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:18
		assert.OK(t, 1 == 2)
		               |
		               false
		
		--- [int] expected
		+++ [int] got
		@@ -1,1 +1,1@@
		-2
		+1
		
		
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:22
		assert.OK(t, a+c == b)
		             ||| |  |
		             ||| |  2
		             ||| false
		             ||3
		             |4
		             1
		
		--- [int] expected
		+++ [int] got
		@@ -1,1 +1,1@@
		-2
		+4
		
		
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:23
		assert.OK(t, (a+c)+a == b)
		              ||| || |  |
		              ||| || |  2
		              ||| || false
		              ||| |1
		              ||| 5
		              ||3
		              |4
		              1
		
		--- [int] expected
		+++ [int] got
		@@ -1,1 +1,1@@
		-2
		+5
		
		
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:24
		assert.OK(t, "foo\nbar" == "bar")
		             |          |
		             |          false
		             "foo
		             bar"
		
		--- [string] expected
		+++ [string] got
		@@ -1,1 +1,2@@
		+foo
		 bar
		
		
--- FAIL: TestStringDiff (0.00s)
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:27
		assert.OK(t, "supersoper" == "supersuper")
		                          |
		                          false
		
		--- [string] expected
		+++ [string] got
		@@ -1,1 +1,1@@
		supers[-u-]{+o+}per
		
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:36
		assert.OK(t, got == expected)
		             |   |  |
		             |   |  "foo
		             |   |  baz
		             |   |  bar2
		             |   |  bar"
		             |   false
		             "foo
		             baz
		             bar
		             bar"
		
		--- [string] expected
		+++ [string] got
		@@ -1,4 +1,4@@
		foo
		baz
		bar[-2-]
		bar
		
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:45
		assert.OK(t, got == expected)
		             |   |  |
		             |   |  "<div>
		             |   |  <div>
		             |   |  foo
		             |   |  </div>
		             |   |  </div>"
		             |   false
		             "<div>
		             bar
		             </div>"
		
		--- [string] expected
		+++ [string] got
		@@ -1,5 +1,3@@
		 <div>
		-<div>
		-foo
		-</div>
		+bar
		 </div>
		
		
--- FAIL: TestMapType (0.00s)
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:50
		assert.OK(t, reflect.DeepEqual(map[string]string{}, map[string]string{"a": "a", k: v}))
		             |                                                                  |  |
		             |                                                                  |  "b------value"
		             |                                                                  "b--------key"
		             false
		
		--- [map[string]string] expected
		+++ [map[string]string] got
		@@ -1,4 +1,1@@
		-{
		-  "b--------key": "b------value",
		-  "a":            "a",
		-}
		+map[string]string{}
		
		
--- FAIL: TestArrayType (0.00s)
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:54
		assert.OK(t, []int{1, 2}[index] == 3)
		                        ||      |
		                        ||      false
		                        |1
		                        2
		
		--- [int] expected
		+++ [int] got
		@@ -1,1 +1,1@@
		-3
		+2
		
		
--- FAIL: TestStructType (0.00s)
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:58
		assert.OK(t, reflect.DeepEqual(struct{ Name string }{foox}, struct{ Name string }{"foo"}))
		             |                                       |
		             |                                       "foo------x"
		             false
		
		--- [struct { Name string }] expected
		+++ [struct { Name string }] got
		@@ -1,3 +1,3@@
		struct { Name string }{
		  Name: "foo{+------x+}",
		}
		
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:59
		assert.OK(t, reflect.DeepEqual(struct{ Name string }{Name: foox}, struct{ Name string }{Name: "foo"}))
		             |                                             |
		             |                                             "foo------x"
		             false
		
		--- [struct { Name string }] expected
		+++ [struct { Name string }] got
		@@ -1,3 +1,3@@
		struct { Name string }{
		  Name: "foo{+------x+}",
		}
		
--- FAIL: TestNestedCallExpr (0.00s)
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:65
		assert.OK(t, rev(rev(rev(true))))
		             |   |   |   |
		             |   |   |   true
		             |   |   false
		             |   true
		             false
		
--- FAIL: TestCallWithNonIdempotentFunc (0.00s)
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:73
		assert.OK(t, incl()+incl() == incl()+incl())
		             |     ||      |  |     ||
		             |     ||      |  |     |4
		             |     ||      |  |     7
		             |     ||      |  3
		             |     ||      false
		             |     |2
		             |     3
		             1
		
		--- [int] expected
		+++ [int] got
		@@ -1,1 +1,1@@
		-7
		+3
		
		
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:74
		assert.OK(t, incl() == incl())
		             |      |  |
		             |      |  6
		             |      false
		             5
		
		--- [int] expected
		+++ [int] got
		@@ -1,1 +1,1@@
		-6
		+5
		
		
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:75
		assert.OK(t, incl() == incl())
		             |      |  |
		             |      |  8
		             |      false
		             7
		
		--- [int] expected
		+++ [int] got
		@@ -1,1 +1,1@@
		-8
		+7
		
		
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:76
		assert.OK(t, incl() == incl())
		             |      |  |
		             |      |  10
		             |      false
		             9
		
		--- [int] expected
		+++ [int] got
		@@ -1,1 +1,1@@
		-10
		+9
		
		
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:77
		assert.OK(t, incl() == incl())
		             |      |  |
		             |      |  12
		             |      false
		             11
		
		--- [int] expected
		+++ [int] got
		@@ -1,1 +1,1@@
		-12
		+11
		
		
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:78
		assert.OK(t, (incl() == incl()) != (incl() == incl()))
		              |      |  |       |   |      |  |
		              |      |  |       |   |      |  16
		              |      |  |       |   |      false
		              |      |  |       |   15
		              |      |  |       false
		              |      |  14
		              |      false
		              13
		
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:84
		assert.OK(t, incl2(incl2(2)) == 10)
		             |     |         |
		             |     |         false
		             |     2
		             4
		
		--- [int] expected
		+++ [int] got
		@@ -1,1 +1,1@@
		-10
		+4
		
		
--- FAIL: TestPkgValue (0.00s)
	assert.go:83: FAIL /.../src/github.com/ToQoz/gopwt/_example/example_test.go:87
		assert.OK(t, sql.ErrNoRows == fmt.Errorf("error"))
		                 |         |  |
		                 |         |  &errors.errorString{s:"error"}
		                 |         false
		                 &errors.errorString{s:"sql: no rows in result set"}
		
		--- [*errors.errorString] expected
		+++ [*errors.errorString] got
		@@ -1,3 +1,3@@
		 &errors.errorString{
		-  s: "error",
		+  s: "sql: no rows in result set",
		 }
		
		
FAIL
FAIL	github.com/ToQoz/gopwt/_example
exit status 1
```

## See Also

- [Why does Go not have assertions?](http://golang.org/doc/faq#assertions)
- http://dontmindthelanguage.wordpress.com/2009/12/11/groovy-1-7-power-assert/
- https://github.com/twada/power-assert
