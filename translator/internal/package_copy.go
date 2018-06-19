package internal

import (
	"os"
)

type File struct {
	SrcPath string
	Path    string
	Mode    os.FileMode
}

// CopyPackage vendoring and ./testdata/... files
func (pkgCtx *PackageContext) CopyPackage() {
	targets := ListTestdataFiles(pkgCtx.Filepath, pkgCtx.Importpath, pkgCtx.SrcDir)
	if pkgCtx.HasVendor {
		targets = append(targets, ListVendorFiles(pkgCtx.Vendor, pkgCtx.Filepath, pkgCtx.Importpath, pkgCtx.SrcDir)...)
	}
	for _, t := range targets {
		pkgCtx.Wg.Add(1)
		func(t File) {
			defer pkgCtx.Wg.Done()

			out, err := os.OpenFile(t.Path, os.O_WRONLY|os.O_CREATE, t.Mode)
			if err != nil {
				pkgCtx.Error = err
				return
			}
			defer out.Close()

			err = CopyFile(t.SrcPath, out)
			if err != nil {
				pkgCtx.Error = err
				return
			}
			return
		}(t)
	}
	if pkgCtx.Error != nil {
		return
	}

	return
}
