package internal

import (
	"bytes"
	"go/parser"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// ReadPackage reads gofiles in package
func (pkgCtx *PackageContext) ReadPackage() {
	// Read & Normalize pkg files
	files, err := ioutil.ReadDir(pkgCtx.Filepath)
	if err != nil {
		pkgCtx.Error = err
		return
	}
	for _, finfo := range files {
		pkgCtx.Wg.Add(1)
		go func(finfo os.FileInfo) {
			defer pkgCtx.Wg.Done()

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
}
