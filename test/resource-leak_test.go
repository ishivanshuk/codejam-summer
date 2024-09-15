package test

import (
	"testing"

	"github.com/mgechev/revive/rule"
)

// Resource leak rule.
func TestResourceLeak(t *testing.T) {
	testRule(t, "resource-leak", &rule.ResourceLeakRule{})
}
