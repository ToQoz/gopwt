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

func newPackageInfo(globalOrLocalImportPath string) (*packageInfo, error) {
	var err error
	var importPath string
	var dirPath string
	var recursive bool

	if strings.HasSuffix(globalOrLocalImportPath, "/...") {
		recursive = true
		globalOrLocalImportPath = strings.TrimSuffix(globalOrLocalImportPath, "/...")
	}

	if globalOrLocalImportPath == "" {
		globalOrLocalImportPath = "."
	}

	if strings.HasPrefix(globalOrLocalImportPath, ".") {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}

		dirPath = filepath.Join(wd, globalOrLocalImportPath)
		if _, err := os.Stat(dirPath); err != nil {
			return nil, err
		}

		importPath, err = findImportPathByPath(dirPath)
		if err != nil {
			return nil, err
		}
	} else {
		importPath = globalOrLocalImportPath

		dirPath, err = findPathByImportPath(importPath)
		if err != nil {
			return nil, err
		}
	}

	return &packageInfo{dirPath: dirPath, importPath: importPath, recursive: recursive}, nil
}

func (pi *packageInfo) ToGoTestArg() string {
	if pi.recursive {
		return filepath.Join(pi.importPath, "...")
	}

	return pi.importPath
}

func findImportPathByPath(path string) (string, error) {
	for _, srcDir := range build.Default.SrcDirs() {
		if path, err := filepath.EvalSymlinks(path); err == nil {
			if srcDir, err := filepath.EvalSymlinks(srcDir); err == nil {
				if strings.HasPrefix(path, srcDir) {
					return strings.TrimPrefix(strings.Replace(path, srcDir, "", 1), "/"), nil
				}
			}
		}
	}

	return "", fmt.Errorf("%s is not found in $GOPATH/src(%q)", path, build.Default.SrcDirs())
}

func findPathByImportPath(importPath string) (string, error) {
	for _, srcDir := range build.Default.SrcDirs() {
		if _, err := os.Stat(filepath.Join(srcDir, importPath)); err == nil {
			return filepath.Join(srcDir, importPath), nil
		}
	}

	return "", fmt.Errorf("package %s is not found in $GOPATH/src(%q)", importPath, build.Default.SrcDirs())
}

func findDeps(importPath, srcDir string) ([]string, error) {
	deps := []string{}

	pkg, err := build.Import(importPath, srcDir, build.AllowBinary)
	if err != nil {
		return nil, err
	}

	for _, imp := range pkg.Imports {
		if imp == importPath {
			continue
		}
		deps = append(deps, imp)
	}

	for _, imp := range pkg.TestImports {
		if imp == importPath {
			continue
		}

		f := false
		for _, arg := range deps {
			if arg == imp {
				f = true
			}
		}

		if !f {
			deps = append(deps, imp)
		}
	}
	return deps, nil
}
