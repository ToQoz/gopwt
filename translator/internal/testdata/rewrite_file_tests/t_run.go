package main

import (
	"testing"

	"github.com/ToQoz/gopwt/assert"
)

func TestTestingRun(t *testing.T) {
	t.Run("t-run", func(tt *testing.T) {
		assert.OK(tt, 1 == 1, "1 is 1")
	})
}
