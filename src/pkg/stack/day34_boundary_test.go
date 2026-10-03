package stack

import (
	"testing"
)

func TestStackBoundaryPaths(t *testing.T) {
	if err := WaitForPods("app=missing", "", 0); err == nil {
		t.Fatal("expected pod wait timeout")
	}
	ctx := &PortForwardContext{}
	ctx.Stop()
	if err := GeneratePM2Config(&StackConfig{}, "/tmp/pm2.js"); err == nil {
		t.Fatal("expected missing dashboard config error")
	}
}
