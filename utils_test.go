package gopwt

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/ToQoz/gopwt/assert"
)

func TestMust(t *testing.T) {
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("must panic if err is given")
			}
		}()

		must(fmt.Errorf("error"))
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("must not panic if err is not given")
			}
		}()

		must(nil)
	}()
}

func TestIsGoFileName(t *testing.T) {
	assert.OK(t, isGoFileName("a.go") == true, "a.go is go file")
	assert.OK(t, isGoFileName("a_test.go") == true, "a_test.go is go file")

	assert.OK(t, isGoFileName(".a.go") == false, ".a.go is not go file")
	assert.OK(t, isGoFileName("_a.go") == false, "_a.go is not go file")
	assert.OK(t, isGoFileName("a.goki") == false, "a.goki is not go file")
	assert.OK(t, isGoFileName("a") == false, "a is not go file")
}

func TestIsTestGoFileName(t *testing.T) {
	assert.OK(t, isTestGoFileName("a_test.go") == true, "a_test.go is test go file")

	assert.OK(t, isTestGoFileName("a.go") == false, "a.go is test go file")
	assert.OK(t, isTestGoFileName(".a_test.go") == false, ".a_test.go is not go file")
	assert.OK(t, isTestGoFileName("_a_test.go") == false, "_a_test.go is not go file")
	assert.OK(t, isTestGoFileName("_a.go") == false, "_a.go is not test go file")
	assert.OK(t, isTestGoFileName("a.goki") == false, "a.goki is not test go file")
	assert.OK(t, isTestGoFileName("a") == false, "a is not test go file")
}

func TestIsBuildableFileName(t *testing.T) {
	assert.OK(t, isBuildableFileName("a.go") == true, "a.go is buildable")

	assert.OK(t, isBuildableFileName("a_test.go") == false, "a_test.go is not buildable")
	assert.OK(t, isBuildableFileName(".a_test.go") == false, ".a_test.go is not buildable")
	assert.OK(t, isBuildableFileName("_a_test.go") == false, "_a_test.go is not buildable")
	assert.OK(t, isBuildableFileName("_a.go") == false, "_a.go is not buildable")
	assert.OK(t, isBuildableFileName("_a.go") == false, "_a.goki is not buildable")
	assert.OK(t, isBuildableFileName("a.goki") == false, "a.goki is not buildable")
	assert.OK(t, isBuildableFileName("a") == false, "a is not buildable")
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
			if isGoFileName(f.Name()) {
				file := filepath.Join(dir, f.Name())
				_, err := parser.ParseFile(fset, file, nil, 0)
				assert.OK(t, err == nil)
			}
		}
	}

	dir = filepath.Join("testdata", "go_files")
	fset = token.NewFileSet()
	parse(dir, fset)
	assert.OK(t, isBuildableFileSet(fset) == true, "testdata/go_files contains go files")

	dir = filepath.Join("testdata", "no_go_files")
	fset = token.NewFileSet()
	parse(dir, fset)
	assert.OK(t, isBuildableFileSet(fset) == false, "testdata/no_go_files contains go files")
}

func TestContainsGoFile(t *testing.T) {
	var files []os.FileInfo
	var err error

	files, err = ioutil.ReadDir("./testdata/go_files")
	assert.Require(t, err == nil)
	assert.OK(t, containsGoFile(files) == true, "./testdata/go_files contains go files")

	files, err = ioutil.ReadDir("./testdata/no_go_files")
	assert.Require(t, err == nil)
	assert.OK(t, containsGoFile(files) == false, "./testdata/no_go_files don't contains go files")
}

func TestContainDirectory(t *testing.T) {
	var files []os.FileInfo
	var err error

	files, err = ioutil.ReadDir("./testdata/dirs")
	assert.Require(t, err == nil)
	assert.OK(t, containsDirectory(files) == true, "./testdata/dirs contains directories")

	files, err = ioutil.ReadDir("./testdata/no_dirs")
	assert.Require(t, err == nil)
	assert.OK(t, containsDirectory(files) == false, "./testdata/no_go_files don't contains directories")
}
