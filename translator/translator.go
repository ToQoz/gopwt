package translator

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/ToQoz/gopwt/translator/internal"
)

// Testdata sets testdata directories that comma separeted
func Testdata(testdata string) {
	internal.Testdata = testdata
}

// WorkingDir sets working dir
func WorkingDir(d string) {
	internal.WorkingDir = d
}

// TermWidth sets term width
func TermWidth(width int) {
	internal.TermWidth = width
}

// Verbose sets verbose or not
func Verbose(v bool) {
	internal.Verbose = v
}

// Translate tlanslates package in given path
func Translate(path string) (gopath, importpath string, err error) {
	var _filepath string

	recursive := false
	if strings.HasSuffix(path, "/...") {
		path = strings.TrimSuffix(path, "/...")
		recursive = true
	}

	gopath, err = ioutil.TempDir(os.TempDir(), "")
	if err != nil {
		return
	}

	importpath, _filepath, err = internal.HandleGlobalOrLocalImportPath(path)
	if err != nil {
		return
	}

	err = internal.Rewrite(gopath, importpath, _filepath, recursive)
	if err != nil {
		return
	}

	return
}
