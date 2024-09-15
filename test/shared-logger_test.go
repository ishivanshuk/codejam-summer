package test

import (
	"testing"

	"github.com/mgechev/revive/rule"
)

// Shared logger rule.
func TestSharedLogger(t *testing.T) {
	testRule(t, "shared-logger", &rule.SharedLogger{})
}
