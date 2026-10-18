package evaluation

import (
	"context"
	"testing"
	"time"

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

func TestOpikConvenienceFunctionsReportMissingCredentials(t *testing.T) {
	t.Setenv("OPIK_API_KEY", "")
	t.Setenv("COMET_API_KEY", "")
	t.Setenv("OPIK_ENABLED", "false")

	checks := []struct {
		name string
		call func() error
	}{
		{name: "upload dataset", call: func() error { _, err := UploadDatasetToOpik(context.Background(), nil); return err }},
		{name: "sync run", call: func() error { _, err := SyncEvaluationRunToOpik(context.Background(), nil, nil); return err }},
		{name: "create trace", call: func() error {
			_, err := CreateTraceInOpik(context.Background(), "trace", "test", time.Time{}, nil, nil)
			return err
		}},
		{name: "update trace", call: func() error { return UpdateTraceInOpik(context.Background(), "trace", time.Time{}, nil, nil, nil) }},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			if err := check.call(); err == nil {
				t.Fatal("expected missing-credentials error")
			}
		})
	}
}
