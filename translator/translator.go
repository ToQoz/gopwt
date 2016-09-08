package translator

import (
	"strings"

	"github.com/ToQoz/gopwt/translator/internal"
)

func Testdata(testdata string) {
	internal.Testdata = testdata
}

func TermWidth(width int) {
	internal.TermWidth = width
}

func Verbose(v bool) {
	internal.Verbose = v
}

func Translate(gopath, path string) (importpath string, err error) {
	var _filepath string

	recursive := false
	if strings.HasSuffix(path, "/...") {
		path = strings.TrimSuffix(path, "/...")
		recursive = true
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
