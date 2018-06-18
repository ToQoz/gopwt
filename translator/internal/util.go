package internal

import (
	"go/ast"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustParse(file *ast.File, err error) *ast.File {
	Must(err)
	return file
}

func Assert(condition bool, msg string) {
	if !condition {
		panic("[assert] " + msg)
	}
}

func CopyFile(path string, out io.Writer) error {
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()
	_, err = io.Copy(out, in)
	return err
}

func IsGoFile(f os.FileInfo) bool {
	name := f.Name()
	return !f.IsDir() && IsGoFileName(name)
}

func IsTestGoFile(f os.FileInfo) bool {
	name := f.Name()
	return !f.IsDir() && IsTestGoFileName(name)
}

func IsGoFileName(name string) bool {
	name = filepath.Base(name)
	return strings.HasSuffix(name, ".go") && !strings.HasPrefix(name, ".") && !strings.HasPrefix(name, "_")
}

func IsTestGoFileName(name string) bool {
	name = filepath.Base(name)
	return strings.HasSuffix(name, "_test.go") && !strings.HasPrefix(name, ".") && !strings.HasPrefix(name, "_")
}

func IsBuildableFileName(name string) bool {
	return IsGoFileName(name) && !IsTestGoFileName(name)
}

func IsBuildableFileSet(s *token.FileSet) bool {
	contains := false
	s.Iterate(func(f *token.File) bool {
		if IsBuildableFileName(f.Name()) {
			contains = true
			return false
		}
		return true
	})
	return contains
}
