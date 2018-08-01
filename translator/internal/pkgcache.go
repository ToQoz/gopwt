package internal

import (
	"crypto/md5"
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	CacheSize = 100
)

type Pkgcache struct {
	PkgRoot    string
	SrcRoot    string
	Importpath string

	PkgPath       string
	PkgcachePath  string
	PkgcacheExist bool
}

func RemoveOldCache() error {
	type fileInfo struct {
		os.FileInfo
		Path string
	}
	var files []fileInfo
	filepath.Walk(CacheDir, func(path string, finfo os.FileInfo, err error) error {
		if finfo.IsDir() {
			return nil
		}

		f := fileInfo{finfo, path}
		modTime := finfo.ModTime()
		inserted := false
		for i, fi := range files {
			if modTime.UnixNano() > fi.ModTime().UnixNano() {
				inserted = true
				files = append(files[:i], append([]fileInfo{f}, files[i:]...)...)
				break
			}
		}
		if !inserted {
			files = append(files, f)
		}
		return nil
	})
	for i, fi := range files {
		p := fi.Path
		if i >= CacheSize {
			if debugLog {
				log.Printf("Delete cache: (%v) %s\n", fi.ModTime(), p)
			}
			if err := os.Remove(p); err != nil {
				return err
			}
		} else {
			if debugLog {
				log.Printf("Keep cache: (%v) %s\n", fi.ModTime(), p)
			}
		}
	}
	return nil
}

// PkgcacheFor returns Pkgcache for given importpath
func PkgcacheFor(hasVendor bool, vendorDir string, importpath string) *Pkgcache {
	goroot := runtime.GOROOT()
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	gopathList := strings.Split(gopath, string(filepath.ListSeparator))

	// Vendoring
	if hasVendor {
		if importpath, ok := RetrieveImportpathFromVendorDir(importpath); ok {
			path := filepath.Join(vendorDir, importpath)
			if _, err := os.Stat(path); err == nil {
				return newPkgcache(filepath.Join(gopathList[0], "pkg"), vendorDir, importpath)
			}
		}
	}

	// GOROOT
	path := filepath.Join(goroot, "src", importpath)
	if _, err := os.Stat(path); err == nil {
		return newPkgcache(filepath.Join(goroot, "pkg"), filepath.Join(goroot, "src"), importpath)
	}

	// GOPATH
	for _, gopath := range gopathList {
		path := filepath.Join(gopath, "src", importpath)
		if _, err := os.Stat(path); err == nil {
			return newPkgcache(filepath.Join(gopath, "pkg"), filepath.Join(gopath, "src"), importpath)
		}
	}

	panic(fmt.Errorf("%s is not found in GOROOT nor GOPATH", importpath))
}

func newPkgcache(pkgRoot, srcRoot string, importpath string) *Pkgcache {
	pkgCacheDir := filepath.Join(CacheDir, importpath)
	c := &Pkgcache{PkgRoot: pkgRoot, SrcRoot: srcRoot, Importpath: importpath}
	pkgDir := filepath.Join(c.SrcRoot, c.Importpath)
	files, err := ioutil.ReadDir(pkgDir)
	if err != nil {
		panic(err)
	}
	hash := []byte{}
	for _, f := range files {
		// NOTE: check *.go except *_test.go
		if !IsGoFile(f) {
			continue
		}
		if IsTestGoFile(f) {
			continue
		}

		f, err := os.Open(filepath.Join(pkgDir, f.Name()))
		if err != nil {
			panic(err)
		}
		defer f.Close()

		h := md5.New()
		if _, err := io.Copy(h, f); err != nil {
			panic(err)
		}
		hash = append(hash, h.Sum(nil)...)
	}
	h := md5.New()
	h.Write(hash)
	rehash := fmt.Sprintf("%x", h.Sum(nil))

	c.PkgPath = filepath.Join(c.PkgRoot, runtime.GOOS+"_"+runtime.GOARCH, c.Importpath+".a")
	c.PkgcachePath = filepath.Join(pkgCacheDir, rehash+".a")

	pkgcache := filepath.Join(pkgCacheDir, rehash+".a")
	_, err = os.Stat(pkgcache)
	c.PkgcacheExist = err == nil

	if c.PkgcacheExist {
		now := time.Now()
		if debugLog {
			log.Printf("Chtimes: (%v) %s", now, c.PkgcachePath)
		}
		if err := os.Chtimes(c.PkgcachePath, now, now); err != nil {
			log.Printf("[WARNING] fail to chtimes cache: (%v) %s", now, c.PkgcachePath)
		}
	}

	return c
}

// Load loads cache
// cp ~/.gopwtcache/pkg/<GOOS>_<GOARCH>/<importpath>/<hash>.a <GOROOT|GOPATH>/pkg/<GOOS>_<GOARCH>/<importpath>.a
func (c *Pkgcache) Load() error {
	in, err := os.Open(c.PkgcachePath)
	if err != nil {
		return err
	}
	defer in.Close()

	if _, err := os.Stat(filepath.Dir(c.PkgPath)); err != nil {
		err = os.MkdirAll(filepath.Dir(c.PkgPath), 0755)
		if err != nil {
			return err
		}
	}
	out, err := os.OpenFile(c.PkgPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// Save saves cache
// cp <GOROOT|GOPATH>/pkg/<GOOS>_<GOARCH>/<importpath>.a ~/.gopwtcache/pkg/<GOOS>_<GOARCH>/<importpath>/<hash>.a
func (c *Pkgcache) Save() error {
	in, err := os.Open(c.PkgPath)
	if err != nil {
		return err
	}
	defer in.Close()

	err = os.MkdirAll(filepath.Dir(c.PkgcachePath), 0755)
	if err != nil {
		return err
	}

	out, err := os.OpenFile(c.PkgcachePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	defer out.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	return err
}
