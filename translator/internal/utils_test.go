package internal_test

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
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
