// +build integration

package gopwt

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ToQoz/gopwt/translator"
)

func init() {

}

func TestIntegrationIssue33(t *testing.T) {
	if err := copyTestdataTo("tdata"); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("tdata")

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("fail to get wd")
	}
	err = os.Chdir(filepath.Join(wd, "tdata", "regression", "issue33"))
	defer os.Chdir(wd)
	if err != nil {
		t.Fatal("fail to chdir")
	}
	tmpGopath, importpath, err := translator.Translate(".")
	buf := bytes.NewBuffer([]byte{})
	err = runTest(tmpGopath, importpath, buf, buf)
	if err != nil {
		fmt.Println(buf.String())
		t.Error("runTest must be pass")
	}
}

func TestIntegrationIssue36(t *testing.T) {
	if err := copyTestdataTo("tdata"); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("tdata")

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("fail to get wd")
	}

	err = os.Chdir(filepath.Join("tdata", "regression", "issue36"))
	defer os.Chdir(wd)
	if err != nil {
		t.Fatal("fail to chdir")
	}

	cmd := exec.Command("dep", "ensure")
	buf := bytes.NewBuffer([]byte{})
	cmd.Stdout = buf
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		t.Error("error on dep ensure: " + buf.String())
	}

	tmpGopath, importpath, err := translator.Translate(".")
	buf = bytes.NewBuffer([]byte{})
	err = runTest(tmpGopath, importpath, buf, buf)
	if err != nil {
		fmt.Println(buf.String())
		t.Error("error on runTest: " + buf.String())
	}
}

func copyTestdataTo(to string) error {
	// dep ensure fails under ./testdata
	// So copy to /tmp
	return filepath.Walk("testdata", func(path string, fInfo os.FileInfo, err error) error {
		if fInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil
		}

		rel, err := filepath.Rel("testdata", path)
		if err != nil {
			return err
		}
		outPath := filepath.Join(to, rel)

		if fInfo.IsDir() {
			di, err := os.Stat(path)
			if err != nil {
				return err
			}
			return os.MkdirAll(outPath, di.Mode())
		}

		in, err := os.OpenFile(path, os.O_RDWR, fInfo.Mode())
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(outPath, os.O_RDWR|os.O_CREATE, fInfo.Mode())
		if err != nil {
			return err
		}
		defer out.Close()

		io.Copy(out, in)
		return nil
	})
}
