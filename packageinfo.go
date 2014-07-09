package main

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

type packageInfo struct {
	dirPath    string
	importPath string
	recursive  bool
}

func (pi *packageInfo) ToGoTestArg() string {
	if pi.recursive {
		return filepath.Join(pi.importPath, "...")
	}

	return pi.importPath
}

func findImportPathByPath(path string) (string, error) {
	for _, srcDir := range build.Default.SrcDirs() {
		if strings.HasPrefix(path, srcDir) {
			return strings.TrimPrefix(strings.Replace(path, srcDir, "", 1), "/"), nil
		}
	}

	return "", fmt.Errorf("%s is not in $GOPATH", path)
}

func findPathByImportPath(importPath string) (string, error) {
	for _, srcDir := range build.Default.SrcDirs() {
		if _, err := os.Stat(filepath.Join(srcDir, importPath)); err == nil {
			return filepath.Join(srcDir, importPath), nil
		}
	}

	return "", fmt.Errorf("package %s is not found in $GOPATH", importPath)
}
