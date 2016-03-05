# gopwt

[![Drone Build Status](https://drone.io/github.com/ToQoz/gopwt/status.png)](https://drone.io/github.com/ToQoz/gopwt/latest)
[![Travis-CI Build Status](https://travis-ci.org/ToQoz/gopwt.svg?branch=master)](https://travis-ci.org/ToQoz/gopwt)

|package|coverage|
|-------|-----|
|gopwt GIT_LATEST_TAG| [![](https://gocover.io/_badge/github.com/toqoz/gopwt?GIT_LATEST_TAG)](https://gocover.io/github.com/toqoz/gopwt)|
|gopwt/assert GIT_LATEST_TAG| [![](https://gocover.io/_badge/github.com/toqoz/gopwt/assert?GIT_LATEST_TAG)](https://gocover.io/github.com/toqoz/gopwt/assert)|
|gopwt/translatedassert GIT_LATEST_TAG| [![](https://gocover.io/_badge/github.com/toqoz/gopwt/translatedassert?GIT_LATEST_TAG)](https://gocover.io/github.com/toqoz/gopwt/translatedassert)|

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
		             | |  "yyy"
		             | false
		             "xxx"

		Assersion messages:
			- a should equal to b

		--- [string] b
		--- [string] a
		@@ -1,1 +1,1@@
		-yyy
		+xxx


FAIL
FAIL    github.com/ToQoz/gopwtexample        0.008s
exit status 1
```

### Update

`go get -u github.com/ToQoz/gopwt/...`
