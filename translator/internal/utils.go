package internal

import (
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func RetrieveImportpathFromVendorDir(path string) (string, bool) {
	segs := strings.SplitN(path, string(filepath.Separator)+"vendor"+string(filepath.Separator), 2)
	if len(segs) < 2 {
		return "", false
	}
	return segs[1], true
}

func ContainsDirectory(files []os.FileInfo) bool {
	for _, f := range files {
		if f.IsDir() {
			return true
		}
	}

	return false
}

func ContainsGoFile(files []os.FileInfo) bool {
	for _, f := range files {
		if isGoFile(f) {
			return true
		}
	}

	return false
}

func IsTestdata(fpath string) bool {
	for _, tdata := range strings.Split(Testdata, ",") {
		if strings.Split(fpath, string(filepath.Separator))[0] == tdata {
			return true
		}
	}
	return false
}

func FindVendor(fpath string, nest int) (string, bool) {
	vdir := fpath
	n := 0
	for {
		v := filepath.Join(vdir, "vendor")
		_, err := os.Stat(v)
		if err == nil {
			return v, true
		}
		n++
		if n > nest {
			break
		}
		vdir = filepath.Dir(fpath)
	}
	return "", false
}

func isGoFile(f os.FileInfo) bool {
	name := f.Name()
	return !f.IsDir() && IsGoFileName(name)
}

func isTestGoFile(f os.FileInfo) bool {
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
