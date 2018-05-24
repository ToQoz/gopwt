package internal

import (
	"os"
	"testing"
)

func TestHandleGlobalOrLocalImportPath(t *testing.T) {
	wd, _ := os.Getwd()
	WorkingDir = wd
	importPath, filepath, _ := HandleGlobalOrLocalImportPath(".")
	if importPath != "github.com/ToQoz/gopwt/translator/internal" {
		t.Errorf("expected=%#v, but got=%#v", "github.com/ToQoz/gopwt/translator/internal", importPath)
	}
	if filepath != wd {
		t.Errorf("expected=%#v, but got=%#v", filepath, wd)
	}

	importPath, filepath, _ = HandleGlobalOrLocalImportPath("")
	if importPath != "github.com/ToQoz/gopwt/translator/internal" {
		t.Errorf("expected=%#v, but got=%#v", "github.com/ToQoz/gopwt/translator/internal", importPath)
	}
	if filepath != wd {
		t.Errorf("expected=%#v, but got=%#v", wd, filepath)
	}

	importPath, filepath, _ = HandleGlobalOrLocalImportPath("github.com/ToQoz/gopwt/translator/internal")
	if importPath != "github.com/ToQoz/gopwt/translator/internal" {
		t.Errorf("expected=%#v, but got=%#v", "github.com/ToQoz/gopwt/translator/internal", importPath)
	}
	if filepath != wd {
		t.Errorf("expected=%#v, but got=%#v", wd, filepath)
	}
}
