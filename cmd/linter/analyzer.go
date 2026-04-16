package main

import (
	"go/ast"
	"go/token"

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

type mainInfo struct {
	pos token.Pos
	end token.Pos
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	var mainFuncs []mainInfo
	if pass.Pkg.Name() == "main" {
		for _, file := range pass.Files {
			for _, decl := range file.Decls {
				if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Name.Name == "main" {
					mainFuncs = append(mainFuncs, mainInfo{
						pos: funcDecl.Pos(),
						end: funcDecl.End(),
					})
				}
			}
		}
	}

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
		(*ast.GoStmt)(nil),
		(*ast.DeferStmt)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		var call *ast.CallExpr
		switch node := n.(type) {
		case *ast.CallExpr:
			call = node
		case *ast.GoStmt:
			call = node.Call
		case *ast.DeferStmt:
			call = node.Call
		default:
			return
		}
		checkCallExpr(pass, call, mainFuncs)
	})

	return nil, nil
}

func isInsideMain(pos token.Pos, mainFuncs []mainInfo) bool {
	for _, m := range mainFuncs {
		if pos > m.pos && pos < m.end {
			return true
		}
	}
	return false
}

func isInsideInit(pass *analysis.Pass, pos token.Pos) bool {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			if funcDecl, ok := decl.(*ast.FuncDecl); ok && funcDecl.Name.Name == "init" {
				if pos > funcDecl.Pos() && pos < funcDecl.End() {
					return true
				}
			}
		}
	}
	return false
}

func checkCallExpr(pass *analysis.Pass, call *ast.CallExpr, mainFuncs []mainInfo) {
	if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" {
		if !isPanicAllowed(pass, call, mainFuncs) {
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
		if !isInsideMain(call.Pos(), mainFuncs) {
			pass.Reportf(call.Pos(), "call to log.%s outside main.main is forbidden", funcName)
		}
		return
	}

	if pkgName == "os" && funcName == "Exit" {
		if !isInsideMain(call.Pos(), mainFuncs) {
			pass.Reportf(call.Pos(), "call to os.Exit outside main.main is forbidden")
		}
		return
	}

	if pkgName == "runtime" && funcName == "Goexit" {
		if !isInsideMain(call.Pos(), mainFuncs) {
			pass.Reportf(call.Pos(), "call to runtime.Goexit outside main.main is forbidden")
		}
		return
	}
}

func isPanicAllowed(pass *analysis.Pass, call *ast.CallExpr, mainFuncs []mainInfo) bool {
	if isInsideMain(call.Pos(), mainFuncs) {
		return true
	}
	if isInsideInit(pass, call.Pos()) {
		return true
	}
	return false
}
