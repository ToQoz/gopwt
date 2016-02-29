package main

import (
	"fmt"
	"github.com/ToQoz/gopwt/assert"
	"strings"
	"testing"
)

func TestMain(t *testing.T) {
	assert.OK(t, strings.Join([]string{"hello", "world"}, ", ") == fmt.Sprintf("%s, %s", "hello", "world"))
}
