package main

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
)

type Package struct {
	Path    string
	Dir     string
	Types   *types.Package
	Structs []*StructInfo
}

type StructInfo struct {
	Name   string
	Obj    *types.TypeName
	Fields []*FieldInfo
}

type FieldInfo struct {
	Name string
	Var  *types.Var
	Tag  string
}

func scanPackages(root string) ([]*Package, error) {
	var pkgs []*Package

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			return nil
		}

		if isTestDir(path) || isVendorDir(path) {
			return filepath.SkipDir
		}

		if !hasGoFiles(path) {
			return nil
		}

		pkg, err := parsePackage(path)
		if err != nil {
			return nil
		}

		if len(pkg.Structs) > 0 {
			pkgs = append(pkgs, pkg)
		}

		return nil
	})

	return pkgs, err
}

func isTestDir(path string) bool {
	return strings.Contains(path, "_test") || strings.HasSuffix(path, "testdata")
}

func isVendorDir(path string) bool {
	return strings.Contains(path, "vendor") || strings.Contains(path, ".git")
}

func hasGoFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
			return true
		}
	}
	return false
}

func parsePackage(dir string) (*Package, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	conf := types.Config{
		Importer: importer.ForCompiler(fset, "source", nil),
		Sizes:    types.SizesFor("gc", "amd64"),
	}

	pkgInfo := &Package{
		Path: dir,
		Dir:  dir,
	}

	for _, astPkg := range pkgs {
		files := make([]*ast.File, 0, len(astPkg.Files))
		for _, f := range astPkg.Files {
			files = append(files, f)
		}

		checkedPkg, err := conf.Check(dir, fset, files, nil)
		if err != nil {
			continue
		}

		pkgInfo.Types = checkedPkg

		for _, file := range files {
			ast.Inspect(file, func(n ast.Node) bool {
				genDecl, ok := n.(*ast.GenDecl)
				if !ok {
					return true
				}

				if genDecl.Doc == nil {
					return true
				}

				hasResetDirective := false
				for _, comment := range genDecl.Doc.List {
					if strings.Contains(comment.Text, "// generate:reset") {
						hasResetDirective = true
						break
					}
				}

				if !hasResetDirective {
					return true
				}

				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}

					_, ok = typeSpec.Type.(*ast.StructType)
					if !ok {
						continue
					}

					obj := checkedPkg.Scope().Lookup(typeSpec.Name.Name)
					if obj == nil {
						continue
					}

					typeName, ok := obj.(*types.TypeName)
					if !ok {
						continue
					}

					structType, ok := typeName.Type().Underlying().(*types.Struct)
					if !ok {
						continue
					}

					structInfo := &StructInfo{
						Name:   typeName.Name(),
						Obj:    typeName,
						Fields: []*FieldInfo{},
					}

					for i := 0; i < structType.NumFields(); i++ {
						field := structType.Field(i)
						if !field.Exported() && field.Name() != "" {
							continue
						}
						structInfo.Fields = append(structInfo.Fields, &FieldInfo{
							Name: field.Name(),
							Var:  field,
							Tag:  structType.Tag(i),
						})
					}

					pkgInfo.Structs = append(pkgInfo.Structs, structInfo)
				}
				return true
			})
		}
	}

	return pkgInfo, nil
}
