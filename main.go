package main

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/singlechecker"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "handleerrorcheck",
	Doc:      "Checks that HandleErr is followed immediately by a return in HTTP Handlers",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	// nodeFilter := []ast.Node{
	// 	(*ast.CallExpr)(nil),
	// 	(*ast.BlockStmt)(nil),
	// 	(*ast.ReturnStmt)(nil),
	// }

	inspector.WithStack(nil, func(n ast.Node, push bool, stack []ast.Node) bool {
		// We only care about CallExpr nodes for our logic
		// if !push {
		// 	return true
		// }

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// Check if the function being called is HandleError
		ident, ok := call.Fun.(*ast.Ident)
		if !ok || ident.Name != "HandleErr" {
			return true
		}

		fmt.Println(ident.Name, len(stack))
		// fmt.Printf("%+v\n", stack[0])
		// Check if the next statement in the block is not a return statement
		if len(stack) < 3 {
			return true
		}

		// not sure what me represents, but it's above the ident
		me := stack[len(stack)-2]
		blk := stack[len(stack)-3]
		// fmt.Printf("%+v\n", blk)
		parentBlock, ok := blk.(*ast.BlockStmt)
		if !ok {
			return true
		}

		for idx, stmt := range parentBlock.List {
			if stmt == me && (idx == len(parentBlock.List)-1 || !isReturnStmt(parentBlock.List[idx+1])) {
				pass.Reportf(call.Pos(), "HandleErr should be immediately followed by a return")
			}
		}
		return true
	})

	return nil, nil
}

func isReturnStmt(n ast.Stmt) bool {
	_, ok := n.(*ast.ReturnStmt)
	return ok
}

func main() {
	analysis.Validate([]*analysis.Analyzer{Analyzer})
	singlechecker.Main(Analyzer)
}
