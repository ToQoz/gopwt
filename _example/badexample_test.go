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
