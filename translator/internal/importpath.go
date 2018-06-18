package internal

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

func HandleGlobalOrLocalImportPath(globalOrLocalImportPath string) (importpath, fpath string, err error) {
	if globalOrLocalImportPath == "" {
		globalOrLocalImportPath = "."
	}

	if strings.HasPrefix(globalOrLocalImportPath, ".") {
		var wd string
		wd, err = os.Getwd()
		if err != nil {
			return
		}

		fpath = filepath.Join(wd, globalOrLocalImportPath)
		if _, err = os.Stat(fpath); err != nil {
			return
		}
		importpath, err = FindImportPathByPath(fpath)
		return
	}

	importpath = globalOrLocalImportPath
	fpath, err = FindPathByImportPath(importpath)
	return
}

func FindImportPathByPath(path string) (string, error) {
	for _, srcDir := range build.Default.SrcDirs() {
		if path, err := filepath.EvalSymlinks(path); err == nil {
			if srcDir, err := filepath.EvalSymlinks(srcDir); err == nil {
				if strings.HasPrefix(path, srcDir) {
					imp := strings.TrimPrefix(strings.Replace(path, srcDir, "", 1), string(filepath.Separator))
					// windows: github.com\ToQoz\gopwt -> github.com/ToQoz/gopwt
					imp = strings.Replace(imp, `\`, "/", -1)
					return imp, nil
				}
			}
		}
	}

	return "", fmt.Errorf("%s is not found in $GOPATH/src(%q)", path, build.Default.SrcDirs())
}

func FindPathByImportPath(importPath string) (string, error) {
	for _, srcDir := range build.Default.SrcDirs() {
		if _, err := os.Stat(filepath.Join(srcDir, importPath)); err == nil {
			return filepath.Join(srcDir, importPath), nil
		}
	}

	return "", fmt.Errorf("package %s is not found in $GOPATH/src(%q)", importPath, build.Default.SrcDirs())
}

func FindDeps(pkg *build.Package) ([]string, error) {
	depMap := map[string]bool{}

	for _, imp := range pkg.Imports {
		depMap[imp] = true
	}
	for _, imp := range pkg.TestImports {
		depMap[imp] = true
	}
	for _, imp := range pkg.XTestImports {
		depMap[imp] = true
	}
	delete(depMap, pkg.ImportPath) // delete self

	deps := make([]string, len(depMap))
	i := 0
	for dep := range depMap {
		deps[i] = dep
		i++
	}
	return deps, nil
}
