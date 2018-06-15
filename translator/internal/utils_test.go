package internal_test

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/ToQoz/gopwt/assert"
	. "github.com/ToQoz/gopwt/translator/internal"
)

func TestMust(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Must panic if err is given")
			}
		}()

		Must(fmt.Errorf("error"))
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Must not panic if err is not given")
			}
		}()

		Must(nil)
	}()
}

func TestMustParse(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Must panic if err is given")
			}
		}()

		MustParse(nil, fmt.Errorf("error"))
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Must not panic if err is not given")
			}
		}()

		MustParse(nil, nil)
	}()
}

func TestAssert(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Must panic if false is given")
			}
		}()

		Assert(false, "")
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Must not panic if true is given")
			}
		}()

		Assert(true, "")
	}()
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

func TestIsTestdata(t *testing.T) {
	assert.OK(t, IsTestdata(filepath.Join("testdata", "x", "a", "a.go")) == true)
	assert.OK(t, IsTestdata(filepath.Join("testdata", "x", "a", "a.txt")) == true)
	assert.OK(t, IsTestdata(filepath.Join("testdata", "x", "a", "a")) == true)
	assert.OK(t, IsTestdata(filepath.Join("testdata", "x.go")) == true)
	assert.OK(t, IsTestdata(filepath.Join("not_testdata", "x.go")) == false)
}

func TestIsGoFileName(t *testing.T) {
	assert.OK(t, IsGoFileName("a.go") == true, "a.go is go file")
	assert.OK(t, IsGoFileName("a_test.go") == true, "a_test.go is go file")

	assert.OK(t, IsGoFileName(".a.go") == false, ".a.go is not go file")
	assert.OK(t, IsGoFileName("_a.go") == false, "_a.go is not go file")
	assert.OK(t, IsGoFileName("a.goki") == false, "a.goki is not go file")
	assert.OK(t, IsGoFileName("a") == false, "a is not go file")
}

func TestIsTestGoFileName(t *testing.T) {
	assert.OK(t, IsTestGoFileName("a_test.go") == true, "a_test.go is test go file")

	assert.OK(t, IsTestGoFileName("a.go") == false, "a.go is test go file")
	assert.OK(t, IsTestGoFileName(".a_test.go") == false, ".a_test.go is not go file")
	assert.OK(t, IsTestGoFileName("_a_test.go") == false, "_a_test.go is not go file")
	assert.OK(t, IsTestGoFileName("_a.go") == false, "_a.go is not test go file")
	assert.OK(t, IsTestGoFileName("a.goki") == false, "a.goki is not test go file")
	assert.OK(t, IsTestGoFileName("a") == false, "a is not test go file")
}

func TestIsBuildableFileName(t *testing.T) {
	assert.OK(t, IsBuildableFileName("a.go") == true, "a.go is buildable")

	assert.OK(t, IsBuildableFileName("a_test.go") == false, "a_test.go is not buildable")
	assert.OK(t, IsBuildableFileName(".a_test.go") == false, ".a_test.go is not buildable")
	assert.OK(t, IsBuildableFileName("_a_test.go") == false, "_a_test.go is not buildable")
	assert.OK(t, IsBuildableFileName("_a.go") == false, "_a.go is not buildable")
	assert.OK(t, IsBuildableFileName("_a.go") == false, "_a.goki is not buildable")
	assert.OK(t, IsBuildableFileName("a.goki") == false, "a.goki is not buildable")
	assert.OK(t, IsBuildableFileName("a") == false, "a is not buildable")
}

func TestContainsBuildable(t *testing.T) {
	var fset *token.FileSet
	var dir string
	var files []os.FileInfo
	var err error

	parse := func(dir string, fset *token.FileSet) {
		files, err = ioutil.ReadDir(dir)
		assert.OK(t, err == nil)
		for _, f := range files {
			if IsGoFileName(f.Name()) {
				file := filepath.Join(dir, f.Name())
				_, err := parser.ParseFile(fset, file, nil, 0)
				assert.OK(t, err == nil)
			}
		}
	}

	dir = filepath.Join("testdata", "go_files")
	fset = token.NewFileSet()
	parse(dir, fset)
	assert.OK(t, IsBuildableFileSet(fset) == true, "testdata/go_files contains go files")

	dir = filepath.Join("testdata", "no_go_files")
	fset = token.NewFileSet()
	parse(dir, fset)
	assert.OK(t, IsBuildableFileSet(fset) == false, "testdata/no_go_files contains go files")
}

func TestContainsGoFile(t *testing.T) {
	var files []os.FileInfo
	var err error

	files, err = ioutil.ReadDir(filepath.Join("testdata", "go_files"))
	assert.Require(t, err == nil)
	assert.OK(t, ContainsGoFile(files) == true, "./testdata/go_files contains go files")

	files, err = ioutil.ReadDir(filepath.Join("testdata", "no_go_files"))
	assert.Require(t, err == nil)
	assert.OK(t, ContainsGoFile(files) == false, "./testdata/no_go_files don't contains go files")
}

func TestContainDirectory(t *testing.T) {
	var files []os.FileInfo
	var err error

	files, err = ioutil.ReadDir(filepath.Join("testdata", "dirs"))
	assert.Require(t, err == nil)
	assert.OK(t, ContainsDirectory(files) == true, "./testdata/dirs contains directories")

	files, err = ioutil.ReadDir(filepath.Join("testdata", "no_dirs"))
	assert.Require(t, err == nil)
	assert.OK(t, ContainsDirectory(files) == false, "./testdata/no_go_files don't contains directories")
}
