package translator

import (
	"io/ioutil"
	"os"

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
	var fpath string

	gopath, err = ioutil.TempDir(os.TempDir(), "")
	if err != nil {
		return
	}

	importpath, fpath, err = internal.HandleGlobalOrLocalImportPath(path)
	if err != nil {
		return
	}

	err = internal.Rewrite(gopath, importpath, fpath)
	return
}
