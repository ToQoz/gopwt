package main

import (
	"github.com/ToQoz/gopwt/assert"
)

func main() {
	assert.OK(nil, "basic" == "basic")

	func() {
		assert.OK(nil, "in func" == "in func")
	}()
}
