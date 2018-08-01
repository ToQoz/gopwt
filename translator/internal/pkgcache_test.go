package internal_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ToQoz/gopwt/assert"
	. "github.com/ToQoz/gopwt/translator/internal"
)

func TestPkgcacheFor(t *testing.T) {
	gopath := strings.Split(os.Getenv("GOPATH"), string(filepath.ListSeparator))[0]

	pkgcache := PkgcacheFor(false, "", "github.com/ToQoz/gopwt")

	assert.OK(t, pkgcache.PkgRoot == filepath.Join(gopath, "pkg"))
	assert.OK(t, pkgcache.SrcRoot == filepath.Join(gopath, "src"))

	osarch := runtime.GOOS + "_" + runtime.GOARCH
	assert.OK(t, pkgcache.PkgPath == filepath.Join(gopath, "pkg", osarch, "github.com", "ToQoz", "gopwt.a"))
	assert.OK(t, filepath.Dir(pkgcache.PkgcachePath) == filepath.Join(CacheDir, "github.com", "ToQoz", "gopwt"))
}

func TestPkgcacheLoad(t *testing.T) {
	temp, err := ioutil.TempDir(os.TempDir(), "")
	defer os.RemoveAll(temp)
	assert.Require(t, err == nil)

	cache := filepath.Join(temp, "cache", "p.a")
	pkg := filepath.Join(temp, "pkg", "p.a")

	os.Mkdir(filepath.Dir(cache), 0755)
	ioutil.WriteFile(cache, []byte("Hello"), 0755)

	c := &Pkgcache{PkgcachePath: cache, PkgPath: pkg}
	err = c.Load()
	assert.Require(t, err == nil)

	data, err := ioutil.ReadFile(pkg)
	assert.Require(t, err == nil)
	assert.OK(t, string(data) == "Hello")
}

func TestPkgcacheSave(t *testing.T) {
	temp, err := ioutil.TempDir(os.TempDir(), "")
	defer os.RemoveAll(temp)
	assert.Require(t, err == nil)

	cache := filepath.Join(temp, "cache", "p.a")
	pkg := filepath.Join(temp, "pkg", "p.a")

	os.Mkdir(filepath.Dir(pkg), 0755)
	ioutil.WriteFile(pkg, []byte("Hello"), 0755)

	c := &Pkgcache{PkgcachePath: cache, PkgPath: pkg}
	err = c.Save()
	assert.Require(t, err == nil)

	data, err := ioutil.ReadFile(cache)
	assert.Require(t, err == nil)
	assert.OK(t, string(data) == "Hello")
}
