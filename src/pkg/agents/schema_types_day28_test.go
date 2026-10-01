package agents

import "testing"

func TestInferTypeRemainingKinds(t *testing.T) {
	cases := []struct {
		name  string
		value interface{}
		want  string
	}{
		{"date", "2026-10-01", "datetime"},
		{"url", "https://example.test", "url"},
		{"text", "plain", "text"},
		{"number", int64(4), "number"},
		{"boolean", true, "boolean"},
		{"array", []interface{}{"x"}, "array"},
		{"object", map[string]interface{}{"x": 1}, "object"},
		{"unknown", struct{}{}, "unknown"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := inferType(tc.value); got != tc.want {
				t.Fatalf("inferType(%#v) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}
