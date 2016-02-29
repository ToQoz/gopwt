package main

import (
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// FIXME: naming
func isGoFile2(name string) bool {
	name = filepath.Base(name)
	return strings.HasSuffix(name, ".go") && !strings.HasPrefix(name, ".") && !strings.HasPrefix(name, "_")
}

func isTestGoFile(name string) bool {
	name = filepath.Base(name)
	return strings.HasSuffix(name, "_test.go") && !strings.HasPrefix(name, ".") && !strings.HasPrefix(name, "_")
}

func isGoFile(f os.FileInfo) bool {
	name := f.Name()
	return !f.IsDir() && isGoFile2(name)
}

func containsDirectory(files []os.FileInfo) bool {
	for _, f := range files {
		if f.IsDir() {
			return true
		}
	}

	return false
}

func containsGoFile(files []os.FileInfo) bool {
	for _, f := range files {
		if isGoFile(f) {
			return true
		}
	}

	return false
}

// FIXME: naming
func containsGoFile2(s *token.FileSet) bool {
	contains := false
	s.Iterate(func(f *token.File) bool {
		if isGoFile2(f.Name()) && !isTestGoFile(f.Name()) {
			contains = true
			return false
		}
		return true
	})
	return contains
}

func getTermCols(fd uintptr) int {
	var sz = struct {
		_    uint16
		cols uint16
		_    uint16
		_    uint16
	}{}
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&sz)))
	return int(sz.cols)
}
