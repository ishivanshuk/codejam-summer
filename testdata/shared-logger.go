// Package test provides tests for the CustomLoggerRule
package test

import (
	"fmt"
	"log"
	mylog "github.com/devrev/shared/log"
)

func noLoggerUsed() {} // OK

func standardLoggerUsed() {
	log.Println("This is a log message") // MATCH /use of non-github.com\/devrev\/shared\/log logger detected/
}

func customLoggerUsed() { // OK
	customLog.Info("This is a custom log message")
}

func customLoggerWithAlias() { // OK
	mylog.Info("This is a custom log message with alias")
}

func multipleLoggersUsed() {
	log.Println("Standard logger") // MATCH /use of non-github.com\/devrev\/shared\/log logger detected/
	customLog.Info("Custom logger") // OK
	fmt.Println("This is fine")
	someOtherLogger.Warn("Another logger") // MATCH /use of non-github.com\/devrev\/shared\/log logger detected/
}

func loggerInNestedFunction() {
	nestedFunc := func() {
		log.Println("Nested log") // MATCH /use of non-github.com\/devrev\/shared\/log logger detected/
	}
	nestedFunc()
}

func customLoggerInTestFile() { // OK
	// Assume this function is in a _test.go file
	log.Println("This should not be flagged in tests")
}

func errorRelatedIdentifiers() { // OK
	err := fmt.Errorf("some error")
	if err != nil {
		err.Error() // This should not be flagged
	}
}

func loggerMethodsWithDifferentNames() {
	someLogger.Debug("Debug message") // MATCH /use of non-github.com\/devrev\/shared\/log logger detected/
	anotherLogger.Info("Info message") // MATCH /use of non-github.com\/devrev\/shared\/log logger detected/
	logger.Warn("Warning message") // MATCH /use of non-github.com\/devrev\/shared\/log logger detected/
	log.Error("Error message") // MATCH /use of non-github.com\/devrev\/shared\/log logger detected/
	myLogger.Fatal("Fatal message") // MATCH /use of non-github.com\/devrev\/shared\/log logger detected/
	customLogger.Panic("Panic message") // MATCH /use of non-github.com\/devrev\/shared\/log logger detected/
}

func allowedLoggingMethods() { // OK
	fmt.Errorf("This is fine")
	t.Errorf("This is also fine in tests")
	e2e.FatalIfError(err, "This is allowed in e2e tests")
}

func directLogImport() {
	import (
		"log" // MATCH /direct import of 'log' package is not allowed. Use github.com\/devrev\/shared\/log instead/
	)
	log.Println("Using directly imported log package")
}

func multipleImports() {
	import (
		"fmt"
		"log" // MATCH /direct import of 'log' package is not allowed. Use github.com\/devrev\/shared\/log instead/
		customlog "github.com/devrev/shared/log"
	)
	log.Println("Standard log") // MATCH /use of non-github.com\/devrev\/shared\/log logger detected/
	customlog.Info("Custom log") // OK
	fmt.Println("This is fine")
}