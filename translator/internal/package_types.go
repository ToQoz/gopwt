package internal

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/importer"
	"log"
	"path/filepath"
	"runtime"

	"go/types"
	"os"
	"os/exec"
	"strings"
)

func (pkgCtx *PackageContext) TypecheckPackage() {
	buildPackage, err := build.Import(pkgCtx.Importpath, pkgCtx.SrcDir, build.AllowBinary)
	if err != nil {
		pkgCtx.Error = err
		return
	}

	deps, err := FindDeps(buildPackage)
	if err != nil {
		pkgCtx.Error = err
		return
	}
	if debugLog {
		log.Printf("deps: %q\n", deps)
	}

	// NOTE: rewrite $importpath to ./vendor/$importpath if $importpath is found in vendor dir
	if pkgCtx.HasVendor {
		pkgToVendor, err := filepath.Rel(pkgCtx.Filepath, pkgCtx.Vendor)
		if err != nil {
			pkgCtx.Error = err
			return
		}
		if !strings.HasPrefix(pkgToVendor, ".") {
			pkgToVendor = "./" + pkgToVendor
		}
		for i, dep := range deps {
			if _, err := os.Stat(filepath.Join(pkgCtx.Vendor, dep)); err == nil {
				dep = pkgToVendor + "/" + dep
			}
			deps[i] = dep
		}
	}

	// NOTE: go install <main pkg> doesn't create <main pkg>.a
	if IsBuildableFileSet(pkgCtx.NormalizedFset) && buildPackage.Name != "main" {
		deps = append(deps, ".")
	}

	pkgCacheDir := filepath.Join(CacheDir, "pkg")
	os.MkdirAll(pkgCacheDir, 0755)

	caches := make([]*Pkgcache, len(deps))
	for i, dep := range deps {
		if dep == "." {
			dep = pkgCtx.Importpath
		}
		caches[i] = PkgcacheFor(pkgCtx.HasVendor, pkgCtx.Vendor, dep)
	}

	for _, c := range caches {
		if c.PkgcacheExist {
			err := c.Load()
			if err != nil {
				pkgCtx.Error = err
				return
			}
			for i, dep := range deps {
				if dep == "." {
					dep = pkgCtx.Importpath
				}
				vimportpath, v := RetrieveImportpathFromVendorDir(dep)
				if v {
					dep = vimportpath
				}
				if dep == c.Importpath {
					deps = append(deps[:i], deps[i+1:]...)
				}
			}
		}
	}

	if len(deps) > 0 {
		install := exec.Command("go", "install")
		install.Dir = pkgCtx.Filepath
		install.Stdout = os.Stdout
		buf := bytes.NewBuffer([]byte{})
		install.Stderr = buf
		if Verbose {
			install.Args = append(install.Args, "-v")
		}
		install.Args = append(install.Args, deps...)
		if debugLog {
			log.Println(strings.Join(install.Args, " "))
		}
		if err := install.Run(); err != nil {
			pkgCtx.Error = fmt.Errorf("[ERROR] %s\n\n%s", strings.Join(install.Args, " "), buf.String())
			return
		}
	} else {
		if debugLog {
			log.Printf("no need to go install")
		}
	}

	if pkgCtx.HasVendor {
		// NOTE: move pkg/$importpath/vendor/$v-importpath to pkg/$v-importpath
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			gopath = build.Default.GOPATH
		}
		gopath = strings.Split(gopath, string(filepath.ListSeparator))[0]
		pd := filepath.Join(gopath, "pkg", runtime.GOOS+"_"+runtime.GOARCH)
		err := filepath.Walk(pd, func(path string, finfo os.FileInfo, err error) error {
			if path == pd {
				return nil
			}

			vimportpath, found := RetrieveImportpathFromVendorDir(path)
			if !found {
				return nil
			}

			outpath := filepath.Join(pd, vimportpath)
			if finfo.IsDir() {
				return os.MkdirAll(outpath, finfo.Mode())
			}
			out, err := os.OpenFile(outpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, finfo.Mode())
			if err != nil {
				return err
			}
			defer out.Close()
			return CopyFile(path, out)
		})
		if err != nil {
			pkgCtx.Error = err
			return
		}
	}

	for _, c := range caches {
		if !c.PkgcacheExist {
			err := c.Save()
			if err != nil {
				pkgCtx.Error = err
				return
			}
		}
	}

	// Assume types from ast.Node
	typesConfig := types.Config{}
	typesConfig.Importer = importer.Default()
	pkg := types.NewPackage(pkgCtx.Importpath, "")
	pkgCtx.TypeInfo = &types.Info{
		Types:      map[ast.Expr]types.TypeAndValue{},
		Defs:       map[*ast.Ident]types.Object{},
		Uses:       map[*ast.Ident]types.Object{},
		Implicits:  map[ast.Node]types.Object{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
		Scopes:     map[ast.Node]*types.Scope{},
		InitOrder:  []*types.Initializer{},
	}

	// WORKARROUND(xtest): if the tests is xtest like `{pkg}_test`, check types only in `{pkg}_test`
	// NOTE: if you check types both `package {pkg}_test` and `package {pkg}`, you'll get `package {pkg}_test; expected {pkg}`
	isXtest := false
	tests := []*ast.File{}
	xtests := []*ast.File{}
	for _, f := range pkgCtx.GoFiles {
		tests = append(tests, f.Normalized)
		if strings.HasSuffix(f.Normalized.Name.Name, "_test") {
			xtests = append(xtests, f.Normalized)
		}
	}
	isXtest = len(xtests) > 0 && len(xtests) != len(tests)

	checker := types.NewChecker(&typesConfig, pkgCtx.NormalizedFset, pkg, pkgCtx.TypeInfo)
	var files []*ast.File
	if isXtest {
		files = xtests
	} else {
		files = tests
	}
	if err := checker.Files(files); err != nil {
		pkgCtx.Error = err
	}
	return
}

func IsTypeConversion(info *types.Info, e *ast.CallExpr) bool {
	if info == nil {
		return false
	}

	funcOrType := DeterminantExprOfIsTypeConversion(e)

	switch funcOrType.(type) {
	case *ast.ChanType, *ast.FuncType, *ast.MapType, *ast.ArrayType, *ast.StructType, *ast.InterfaceType:
		return true
	case *ast.Ident:
		id := funcOrType.(*ast.Ident)

		if t, ok := info.Types[id]; ok {
			return t.IsType()
		}

		if o := info.ObjectOf(id); o != nil {
			switch o.(type) {
			case *types.TypeName:
				return true
			default:
				return false
			}
		}
	}

	return false
}
