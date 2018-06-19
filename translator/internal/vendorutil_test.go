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

func TestListVendorFiles(t *testing.T) {
	root, err := ioutil.TempDir(os.TempDir(), "")
	assert.Require(t, err == nil)

	tempsrc := filepath.Join(root, "test-list-testdata-files")
	os.MkdirAll(filepath.Join(tempsrc, "github.com", "ToQoz", "gopwt", "translator", "internal", "testdata", "with_vendor"), 0755)

	gopaths := strings.Split(os.Getenv("GOPATH"), string(filepath.ListSeparator))
	gopath := gopaths[len(gopaths)-1]

	var files []string
	pkgdir := filepath.Join(gopath, "src", "github.com", "ToQoz", "gopwt", "translator", "internal", "testdata", "with_vendor")
	vendor := filepath.Join(pkgdir, "vendor")
	for _, f := range ListVendorFiles(vendor, pkgdir, "github.com/ToQoz/gopwt/translator/internal/testdata/with_vendor", tempsrc) {
		files = append(files, f.Path)
	}
	assert.OK(t, len(files) == 6)
}

func TestRetrieveImportpathFromVendorDir(t *testing.T) {
	{
		vpkg, hasVendor := RetrieveImportpathFromVendorDir(filepath.Join("pkg", "path"))
		assert.OK(t, hasVendor == false)
		assert.OK(t, vpkg == "")
	}

	{
		vpkg, hasVendor := RetrieveImportpathFromVendorDir(filepath.Join("pkg", "vendor", "v-pkg", "sub"))
		assert.OK(t, hasVendor == true)
		assert.OK(t, vpkg == filepath.Join("v-pkg", "sub"))
	}
}

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
