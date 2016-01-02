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
