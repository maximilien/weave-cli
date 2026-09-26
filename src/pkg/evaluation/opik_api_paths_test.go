package evaluation

import (
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/llm"
)

func TestOpikConfigurationErrors(t *testing.T) {
	if _, err := NewOpikAPIClient(nil); err == nil {
		t.Fatal("expected missing config error")
	}
	if _, err := NewOpikAPIClient(&llm.OpikConfig{}); err == nil {
		t.Fatal("expected missing API key error")
	}
}
