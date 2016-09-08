package main

import (
	"testing"

	"github.com/ToQoz/gopwt/assert"
)

func TestSimple(t *testing.T) {
	func() {
		assert.OK(t, 1 == 1, "1 is 1")
	}()
}
