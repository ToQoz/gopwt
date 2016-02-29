package main

import (
	"github.com/ToQoz/gopwt/assert"
	"testing"
)

func TestSimple(t *testing.T) {
	func() {
		assert.OK(t, 1 == 1, "1 is 1")
	}()
}
