package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

// ListVendor lists all vendored files
func ListVendorFiles(vendor, pkgDir, importPath, tempGoSrcDir string) []File {
	var targets []File
	filepath.Walk(vendor, func(path string, finfo os.FileInfo, err error) error {
		if path == pkgDir || finfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil
		}

		pathFromPkgDir, err := filepath.Rel(pkgDir, path)
		Assert(err == nil, fmt.Sprintf("filepath.Rel(%s, %s) must be ok", pkgDir, path))
		outpath := filepath.Join(tempGoSrcDir, importPath, pathFromPkgDir)

		if finfo.IsDir() {
			return os.Mkdir(outpath, finfo.Mode())
		}

		targets = append(targets, File{
			Path:    outpath,
			SrcPath: path,
			Mode:    finfo.Mode(),
		})
		return nil
	})
	return targets
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
		vdir = filepath.Dir(vdir)
	}
	return "", false
}
