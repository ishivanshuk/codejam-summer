package rule

import (
	"go/ast"
	"go/token"

	"github.com/mgechev/revive/lint"
)

// ResourceLeakRule lints for resource leaks, specifically ensuring that
// all network and file connections are properly closed.
type ResourceLeakRule struct{}

// Name returns the rule name.
func (*ResourceLeakRule) Name() string {
	return "resource-leak"
}

// Apply applies the rule to the given file.
func (r *ResourceLeakRule) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	if file.Pkg.IsMain() || file.IsTest() {
		return nil
	}

	const (
		message  = "resource opened should be closed"
		category = "resource-management"
	)

	var failures []lint.Failure

	// Track open connections and ensure they are closed properly
	openCalls := map[token.Pos]*ast.CallExpr{}
	closeCalls := map[token.Pos]bool{}

	ast.Inspect(file.AST, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			// Track 'Open' and 'Dial' calls as potential resource opening
			if isOpenCall(x) {
				openCalls[x.Pos()] = x
			}
			// Track 'Close' calls as resource closing
			if isCloseCall(x) {
				closeCalls[x.Pos()] = true
			}
		}
		return true
	})

	for pos, call := range openCalls {
		// Check if there's a corresponding 'Close' call in the file
		if _, found := closeCalls[pos]; !found {
			failures = append(failures, lint.Failure{
				Failure:   message,
				Category:  category,
				Node:      call,
				Confidence: 1,
			})
		}
	}

	return failures
}

// isOpenCall checks if the given CallExpr represents a resource opening call.
func isOpenCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		switch sel.Sel.Name {
		case "Open", "Dial":
			return true
		}
	}
	return false
}

// isCloseCall checks if the given CallExpr represents a resource closing call.
func isCloseCall(call *ast.CallExpr) bool {
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		return sel.Sel.Name == "Close"
	}
	return false
}
