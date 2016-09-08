package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ToQoz/gopwt/assert"
)

func TestMain(t *testing.T) {
	assert.OK(t, strings.Join([]string{"hello", "world"}, ", ") == fmt.Sprintf("%s, %s", "hello", "world"))
}
