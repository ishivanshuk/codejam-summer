package rule

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/mgechev/revive/lint"
)

// ResourceLeakRule lints for potential resource leaks.
type ResourceLeakRule struct{}

// Apply applies the rule to given file.
func (*ResourceLeakRule) Apply(file *lint.File, _ lint.Arguments) []lint.Failure {
	var failures []lint.Failure

	fileAst := file.AST
	walker := lintResourceLeak{
		file:    file,
		fileAst: fileAst,
		onFailure: func(failure lint.Failure) {
			failures = append(failures, failure)
		},
	}

	ast.Walk(walker, fileAst)

	return failures
}

// Name returns the rule name.
func (*ResourceLeakRule) Name() string {
	return "resource-leak"
}

type lintResourceLeak struct {
	file      *lint.File
	fileAst   *ast.File
	onFailure func(lint.Failure)
}

func (w lintResourceLeak) Visit(n ast.Node) ast.Visitor {
	switch stmt := n.(type) {
	case *ast.FuncDecl:
		w.checkFunction(stmt)
	}
	return w
}

func (w lintResourceLeak) checkFunction(fn *ast.FuncDecl) {
	ast.Inspect(fn, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			for _, expr := range stmt.Rhs {
				if call, ok := expr.(*ast.CallExpr); ok {
					if fun, ok := call.Fun.(*ast.SelectorExpr); ok {
						if isResourceOpener(fun) {
							if !hasMatchingDeferOrDirectClosure(stmt.Pos(), fn) {
								w.onFailure(lint.Failure{
									Confidence: 1,
									Node:       stmt,
									Category:   "resource-management",
									Failure:    fmt.Sprintf("check if the resource is closed: %v", w.file.Render(stmt)),
								})
							}
						}
					}
				}
			}
		}
		return true
	})
}

func isResourceOpener(fun *ast.SelectorExpr) bool {
	// openers used by well known packages
	openers := map[string][]string{
		"os":      {"Open", "Create", "ReadFile", "WriteFile"},
		"net":     {"Dial", "Listen"},
		"sql":     {"Open"},
		"bufio":   {"NewReader", "NewWriter", "NewReadWriter"},
		"archive": {"NewReader", "NewWriter"},
		// Add more packages and functions as needed
	}

	if ident, ok := fun.X.(*ast.Ident); ok {
		if funcs, exists := openers[ident.Name]; exists {
			for _, f := range funcs {
				if fun.Sel.Name == f {
					return true
				}
			}
		}
	}
	genericOpeners := []string{
		"init",
		"open",
		"start",
		"dial",
	}

	// Check if the function's name contains any of the generic openers.
	methodName := strings.ToLower(fun.Sel.Name)
	for _, substr := range genericOpeners {
		if containsSubstring(methodName, substr) {
			return true
		}
	}

	return false
}

// Helper function to check if a string contains a substring.
func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}

func hasMatchingDeferOrDirectClosure(pos token.Pos, fn *ast.FuncDecl) bool {
	var hasDefer bool
	var hasDirectClosure bool

	ast.Inspect(fn, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.DeferStmt:
			if stmt.Pos() > pos {
				if call, ok := stmt.Call.Fun.(*ast.SelectorExpr); ok {
					if isResourceCloser(call) {
						hasDefer = true
						return false
					}
				}
			}
		case *ast.ExprStmt:
			if call, ok := stmt.X.(*ast.CallExpr); ok {
				if fun, ok := call.Fun.(*ast.SelectorExpr); ok {
					if isResourceCloser(fun) && call.Pos() > pos {
						hasDirectClosure = true
						return false
					}
				}
			}
		}
		return true
	})

	return hasDefer || hasDirectClosure
}

func isResourceCloser(fun *ast.SelectorExpr) bool {
	// Closers used by well known packages, add closers in same case
	closers := []string{
		"close",
		"stop",
		"shutdown",
		"terminate",
		"abort",
		"exit",
		"destroy",
		"finish",
	}

	if _, ok := fun.X.(*ast.Ident); ok {
		for _, c := range closers {
			if strings.ToLower(fun.Sel.Name) == c {
				return true
			}
		}
	}
	return false
}
