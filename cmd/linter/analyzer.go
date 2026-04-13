package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const doc = "checker detects forbidden calls: panic, log.Fatal, os.Exit"

var Analyzer = &analysis.Analyzer{
	Name:     "checker",
	Doc:      doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
		(*ast.GoStmt)(nil),
		(*ast.DeferStmt)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		switch call := n.(type) {
		case *ast.CallExpr:
			checkCallExpr(pass, call)
		case *ast.GoStmt:
			checkCallExpr(pass, call.Call)
		case *ast.DeferStmt:
			checkCallExpr(pass, call.Call)
		}
	})

	return nil, nil
}

func checkCallExpr(pass *analysis.Pass, call *ast.CallExpr) {

	if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" {
		if !isPanicAllowed(pass, call) {
			pass.Reportf(call.Pos(), "use of panic is forbidden")
		}
		return
	}

	fun, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	xIdent, ok := fun.X.(*ast.Ident)
	if !ok {
		return
	}

	funcName := fun.Sel.Name
	pkgName := xIdent.Name

	if pkgName == "" {
		return
	}

	if pkgName == "log" && (funcName == "Fatal" || funcName == "Fatalf" || funcName == "Fatalln") {
		if !isInMainMain(pass) {
			pass.Reportf(call.Pos(), "call to log.%s outside main.main is forbidden", funcName)
		}
		return
	}

	if pkgName == "os" && funcName == "Exit" {
		if !isInMainMain(pass) {
			pass.Reportf(call.Pos(), "call to os.Exit outside main.main is forbidden")
		}
		return
	}

	if pkgName == "runtime" && funcName == "Goexit" {
		if !isInMainMain(pass) {
			pass.Reportf(call.Pos(), "call to runtime.Goexit outside main.main is forbidden")
		}
		return
	}
}

func isInMainMain(pass *analysis.Pass) bool {
	if pass.Pkg.Name() != "main" {
		return false
	}

	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if ok && funcDecl.Name.Name == "main" {
				return true
			}
		}
	}
	return false
}

func isPanicAllowed(pass *analysis.Pass, call *ast.CallExpr) bool {

	if isInMainMain(pass) {
		return true
	}

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if ok && funcDecl.Name.Name == "init" {
				if call.Pos() > funcDecl.Pos() && call.Pos() < funcDecl.End() {
					return true
				}
			}
		}
	}

	return false
}
