package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ListTestdata lists all <pkgDir>/testdata/**/* files
func ListTestdataFiles(pkgDir string, importPath string, tempGoSrcDir string) []File {
	var targets []File
	filepath.Walk(pkgDir, func(path string, finfo os.FileInfo, err error) error {
		if path == pkgDir || finfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil
		}

		pathFromPkgDir, err := filepath.Rel(pkgDir, path)
		Assert(err == nil, fmt.Sprintf("filepath.Rel(%s, %s) must be ok", pkgDir, path))
		outpath := filepath.Join(tempGoSrcDir, importPath, pathFromPkgDir)

		if finfo.IsDir() {
			if IsTestdata(pathFromPkgDir) {
				return os.Mkdir(outpath, finfo.Mode())
			}
			return filepath.SkipDir
		}

		if IsTestdata(pathFromPkgDir) {
			targets = append(targets, File{
				Path:    outpath,
				SrcPath: path,
				Mode:    finfo.Mode(),
			})
		}
		return nil
	})
	return targets
}

func IsTestdata(fpath string) bool {
	for _, tdata := range strings.Split(Testdata, ",") {
		if strings.Split(fpath, string(filepath.Separator))[0] == tdata {
			return true
		}
	}
	return false
}
