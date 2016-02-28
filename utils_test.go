package main

import (
	"fmt"
	"github.com/ToQoz/gopwt/assert"
	"io/ioutil"
	"os"
	"testing"
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

func TestIsGoFile2(t *testing.T) {
	assert.OK(t, isGoFile2("a.go") == true, "a.go is go file")
	assert.OK(t, isGoFile2("a_test.go") == true, "a_test.go is go file")

	assert.OK(t, isGoFile2(".a.go") == false, ".a.go is not go file")
	assert.OK(t, isGoFile2("_a.go") == false, "_a.go is not go file")
	assert.OK(t, isGoFile2("a.goki") == false, "a.goki is not go file")
	assert.OK(t, isGoFile2("a") == false, "a is not go file")
}

func TestIsTestGoFile(t *testing.T) {
	assert.OK(t, isTestGoFile("a_test.go") == true, "a_test.go is test go file")

	assert.OK(t, isTestGoFile("a.go") == false, "a.go is test go file")
	assert.OK(t, isGoFile2(".a_test.go") == false, ".a_test.go is not go file")
	assert.OK(t, isGoFile2("_a_test.go") == false, "_a_test.go is not go file")
	assert.OK(t, isTestGoFile("_a.go") == false, "_a.go is not test go file")
	assert.OK(t, isTestGoFile("a.goki") == false, "a.goki is not test go file")
	assert.OK(t, isTestGoFile("a") == false, "a is not test go file")
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
