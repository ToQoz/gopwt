package internal

import (
	"bytes"
	"go/parser"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type File struct {
	SrcPath string
	Path    string
	Mode    os.FileMode
}

// CopyPackage copy package
func (pkgCtx *PackageContext) CopyPackage() {
	wg := &sync.WaitGroup{}

	// Copy vendoring dir and ./testdata/...
	targets := ListTestdataFiles(pkgCtx.Filepath, pkgCtx.Importpath, pkgCtx.SrcDir)
	if pkgCtx.HasVendor {
		targets = append(targets, ListVendorFiles(pkgCtx.Vendor, pkgCtx.Filepath, pkgCtx.Importpath, pkgCtx.SrcDir)...)
	}
	for _, t := range targets {
		wg.Add(1)
		func(t File) {
			defer wg.Done()

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

	// Read & Normalize pkg files
	files, err := ioutil.ReadDir(pkgCtx.Filepath)
	if err != nil {
		pkgCtx.Error = err
		return
	}
	for _, finfo := range files {
		wg.Add(1)
		go func(finfo os.FileInfo) {
			defer wg.Done()

			if finfo.IsDir() {
				return
			}

			if !IsGoFile(finfo) {
				return
			}

			path := filepath.Join(pkgCtx.Filepath, finfo.Name())
			pathFromPkgDir, err := filepath.Rel(pkgCtx.Filepath, path)
			if err != nil {
				pkgCtx.Error = err
				return
			}
			outpath := filepath.Join(pkgCtx.SrcDir, pkgCtx.Importpath, pathFromPkgDir)

			in, err := os.Open(path)
			if err != nil {
				pkgCtx.Error = err
				return
			}
			defer in.Close()

			file, err := parser.ParseFile(pkgCtx.OriginalFset, path, in, 0)
			if err != nil {
				pkgCtx.Error = err
				return
			}

			out := bytes.NewBuffer([]byte{})
			defer func() {
				pkgCtx.GoFiles = append(pkgCtx.GoFiles, &GoFile{
					Path:       outpath,
					Original:   file,
					Normalized: MustParse(parser.ParseFile(pkgCtx.NormalizedFset, outpath, out, 0)),
					Mode:       finfo.Mode(),
					Data:       out,
				})
			}()
			if !IsTestGoFileName(path) {
				in.Seek(0, SeekStart)
				_, err := io.Copy(out, in)
				if err != nil {
					pkgCtx.Error = err
				}
				return
			}
			err = NormalizeFile(file, in, out)
			if err != nil {
				pkgCtx.Error = err
			}
			return
		}(finfo)
	}

	wg.Wait()
	return
}
