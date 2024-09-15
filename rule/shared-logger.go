package rule

import (
	"fmt"
	"go/ast"
	"strings"
	"sync"

	"github.com/mgechev/revive/lint"
)

// SharedLogger lints usage of non-custom loggers.
type SharedLogger struct {
	sync.Mutex
	configured     bool
	allowedPackage string
}

func (r *SharedLogger) configure(args lint.Arguments) {
	r.Lock()
	defer r.Unlock()

	if r.configured {
		return
	}
	r.configured = true

	if len(args) > 0 {
		options, ok := args[0].(map[string]interface{})
		if !ok {
			panic(fmt.Errorf("invalid rule configuration for %s", r.Name()))
		}
		if allowedPackage, ok := options["allowedPackage"].(string); ok {
			r.allowedPackage = allowedPackage
		}
	}

	if r.allowedPackage == "" {
		r.allowedPackage = "github.com/devrev/shared/log"
	}
}

// Apply applies the rule to given file.
func (r *SharedLogger) Apply(file *lint.File, args lint.Arguments) []lint.Failure {
	r.configure(args)

	var failures []lint.Failure

	walker := &sharedLoggerWalker{
		file:           file,
		allowedPackage: r.allowedPackage,
		customImport:   make(map[string]logAlias),
		onFailure: func(failure lint.Failure) {
			failures = append(failures, failure)
		},
	}

	ast.Walk(walker, file.AST)

	return failures
}

// Name returns the rule name.
func (*SharedLogger) Name() string {
	return "custom-logger"
}

type logAlias struct {
	aliasName string
	fileName  string
}

type sharedLoggerWalker struct {
	file           *lint.File
	allowedPackage string
	customImport   map[string]logAlias
	onFailure      func(lint.Failure)
}

func (w *sharedLoggerWalker) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.ImportSpec:
		w.checkImport(n)
	case *ast.SelectorExpr:
		w.checkLoggerUsage(n)
	}
	return w
}

func (w *sharedLoggerWalker) checkImport(spec *ast.ImportSpec) {
	if spec.Path.Value == `"log"` {
		w.onFailure(lint.Failure{
			Confidence: 1,
			Node:       spec,
			Failure:    fmt.Sprintf("direct import of 'log' package is not allowed. Use %s instead", w.allowedPackage),
		})
	}
	if spec.Path.Value == `"`+w.allowedPackage+`"` {
		if spec.Name != nil {
			w.customImport[spec.Name.Name] = logAlias{aliasName: spec.Name.Name, fileName: w.file.Name}
		} else {
			w.customImport["log"] = logAlias{aliasName: "log", fileName: w.file.Name}
		}
	}
}

func (w *sharedLoggerWalker) checkLoggerUsage(sel *ast.SelectorExpr) {
	if ident, ok := sel.X.(*ast.Ident); ok {
		if strings.HasSuffix(w.file.Name, "_test.go") {
			return
		}
		if ident.Name == "zap" || ident.Name == "mlRunLog" {
			return
		}
		if (ident.Name == "fmt" || ident.Name == "t" || ident.Name == "e2e") && (sel.Sel.Name == "Errorf" || sel.Sel.Name == "FatalIfError" || sel.Sel.Name == "Fatalf") {
			return
		}
		if !w.checkAlias(ident.Name) && isLoggerMethod(ident.Name, sel.Sel.Name) {
			w.onFailure(lint.Failure{
				Confidence: 1,
				Node:       sel,
				Failure:    fmt.Sprintf("use of non-%s logger detected", w.allowedPackage),
			})
		}
	}
}

func (w *sharedLoggerWalker) checkAlias(identName string) bool {
	for _, alias := range w.customImport {
		if alias.aliasName == identName && alias.fileName == w.file.Name {
			return true
		}
	}
	return false
}

func isLoggerMethod(identName string, name string) bool {
	logMethods := []string{"Debug", "Info", "Warn", "Error", "Fatal", "Panic"}
	for _, method := range logMethods {
		if method == "Error" && (strings.Contains(strings.ToLower(identName), "err") || strings.Contains(strings.ToLower(identName), "status") || strings.Contains(strings.ToLower(identName), "api")) {
			return false
		}
		if strings.HasPrefix(name, method) {
			return true
		}
	}
	return false
}