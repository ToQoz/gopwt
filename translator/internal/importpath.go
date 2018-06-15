package internal

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

func HandleGlobalOrLocalImportPath(globalOrLocalImportPath string) (importpath, _filepath string, err error) {
	if globalOrLocalImportPath == "" {
		globalOrLocalImportPath = "."
	}

	if strings.HasPrefix(globalOrLocalImportPath, ".") {
		wd, e := os.Getwd()
		if e != nil {
			err = e
			return
		}

		_filepath = filepath.Join(wd, globalOrLocalImportPath)
		if _, err = os.Stat(_filepath); err != nil {
			return
		}

		importpath, err = FindImportPathByPath(_filepath)
		if err != nil {
			return
		}
	} else {
		importpath = globalOrLocalImportPath

		_filepath, err = FindPathByImportPath(importpath)
		if err != nil {
			return
		}
	}

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

func FindDeps(importPath, srcDir string) ([]string, error) {
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

	for _, imp := range pkg.XTestImports {
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
