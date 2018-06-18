package internal_test

import (
	"path/filepath"
	"testing"

	"github.com/ToQoz/gopwt/assert"
	. "github.com/ToQoz/gopwt/translator/internal"
)

func TestIsTestdata(t *testing.T) {
	assert.OK(t, IsTestdata(filepath.Join("testdata", "x", "a", "a.go")) == true)
	assert.OK(t, IsTestdata(filepath.Join("testdata", "x", "a", "a.txt")) == true)
	assert.OK(t, IsTestdata(filepath.Join("testdata", "x", "a", "a")) == true)
	assert.OK(t, IsTestdata(filepath.Join("testdata", "x.go")) == true)
	assert.OK(t, IsTestdata(filepath.Join("not_testdata", "x.go")) == false)
}
