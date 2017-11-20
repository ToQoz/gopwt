package translator

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/ToQoz/gopwt/translator/internal"
)

func Testdata(testdata string) {
	internal.Testdata = testdata
}

func WorkingDir(d string) {
	internal.WorkingDir = d
}

func TermWidth(width int) {
	internal.TermWidth = width
}

func Verbose(v bool) {
	internal.Verbose = v
}

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
