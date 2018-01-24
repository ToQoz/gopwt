// Copyright 2014-2016 Takatoshi Matsumoto. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package gopwt brings PowerAssert to golang

Installation
	$ go get github.com/ToQoz/gopwt/...

Usage
	$ cd $GOPATH/src/your-go-project-path
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
	--- FAIL: TestFoo (0.00s)
		assert.go:61: FAIL /var/folders/f8/0gm3xlgn1q12_zt7kxmzfj480000gn/T/090346829/src/github.com/ToQoz/gopwt/foo/main_test.go:11
			assert.OK(t, a == b)
				     | |  |
				     | |  "b"
				     | false
				     "a"

	FAIL
	FAIL    github.com/ToQoz/gopwt/foo      0.006s
	exit status 1
*/
package gopwt
