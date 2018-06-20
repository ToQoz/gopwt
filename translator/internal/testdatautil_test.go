package internal_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ToQoz/gopwt/assert"
	. "github.com/ToQoz/gopwt/translator/internal"
)

func TestIsTestdata(t *testing.T) {
	assert.OK(t, IsTestdata(filepath.Join("testdata", "x", "a", "a.go")) == true)
	assert.OK(t, IsTestdata(filepath.Join("testdata", "x", "a", "a.txt")) == true)
	assert.OK(t, IsTestdata(filepath.Join("testdata", "x", "a", "a")) == true)
	assert.OK(t, IsTestdata(filepath.Join("testdata", "x.go")) == true)
	assert.OK(t, IsTestdata(filepath.Join("not_testdata", "x.go")) == false)
}

func TestListTestdataFiles(t *testing.T) {
	temp, err := ioutil.TempDir(os.TempDir(), "")
	defer os.RemoveAll(temp)

	tempsrc := filepath.Join(temp, "test-list-testdata-files")
	os.MkdirAll(filepath.Join(tempsrc, "github.com", "ToQoz", "gopwt", "translator", "internal"), 0755)

	wd, err := os.Getwd()
	assert.Require(t, err == nil)
	var files []string
	for _, f := range ListTestdataFiles(wd, "github.com/ToQoz/gopwt/translator/internal", tempsrc) {
		files = append(files, f.Path)
	}

	count := 0
	filepath.Walk(wd, func(path string, finfo os.FileInfo, err error) error {
		if finfo.IsDir() {
			return nil
		}
		if strings.Contains(path, string(filepath.Separator)+"testdata"+string(filepath.Separator)) {
			count++
		}
		return nil
	})
	assert.OK(t, len(files) == count)
}
