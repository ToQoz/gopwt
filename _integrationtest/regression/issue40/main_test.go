// https://github.com/ToQoz/gopwt/issues/40
package main

import (
	"flag"
	"os"

	. "github.com/ToQoz/gopwt"
	. "github.com/ToQoz/gopwt/assert"

	"testing"
)

type myint int

const two myint = 2

func TestMain(m *testing.M) {
	flag.Parse()
	Empower()
	os.Exit(m.Run())
}

func TestFoo(t *testing.T) {
	a := myint(6)
	OK(t, a == 3*two)
}
