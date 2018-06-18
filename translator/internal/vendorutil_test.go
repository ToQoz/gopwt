package internal_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ToQoz/gopwt/assert"
	. "github.com/ToQoz/gopwt/translator/internal"
)

func TestFindVendor(t *testing.T) {
	tests := []struct {
		pkg            string
		vendor         string
		expectedVendor string
	}{
		{
			pkg:            filepath.Join("a", "b", "c", "d"),
			vendor:         filepath.Join("a", "b", "c", "d", "vendor"),
			expectedVendor: filepath.Join("a", "b", "c", "d", "vendor"),
		},
		{
			pkg:            filepath.Join("a", "b", "c", "d"),
			vendor:         filepath.Join("a", "b", "c", "vendor"),
			expectedVendor: filepath.Join("a", "b", "c", "vendor"),
		},
		{
			pkg:            filepath.Join("a", "b", "c", "d"),
			vendor:         filepath.Join("a", "b", "vendor"),
			expectedVendor: filepath.Join("a", "b", "vendor"),
		},
		{
			pkg:            filepath.Join("a", "b", "c", "d"),
			vendor:         filepath.Join("a", "vendor"),
			expectedVendor: filepath.Join("a", "vendor"),
		},
		{
			pkg:            filepath.Join("a", "b", "c", "d"),
			vendor:         filepath.Join("vendor"),
			expectedVendor: "",
		},
	}
	root, err := ioutil.TempDir(os.TempDir(), "")
	for i, test := range tests {
		src := filepath.Join(root, "src-test-find-vendor-"+strconv.Itoa(i))

		assert.Require(t, err == nil)
		if test.vendor != "" {
			os.MkdirAll(filepath.Join(src, test.vendor), 0755)
		}

		vendor, hasVendor := FindVendor(filepath.Join(src, test.pkg), strings.Count(test.pkg, string(filepath.Separator)))
		assert.OK(t, hasVendor == (test.expectedVendor != ""))
		if test.expectedVendor != "" {
			rel, err := filepath.Rel(src, vendor)
			assert.Require(t, err == nil)
			assert.OK(t, rel == test.expectedVendor)
		}

		err = os.RemoveAll(src)
		assert.Require(t, err == nil)
	}
}
