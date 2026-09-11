// SPDX-License-Identifier: MIT
// Copyright (c) 2026 dr.max

package serve

import (
	"os"
	"syscall"
	"testing"
)

func TestServeCommandContract(t *testing.T) {
	if ServeCmd.Use != "serve" || ServeCmd.RunE == nil {
		t.Fatalf("ServeCmd = %+v", ServeCmd)
	}
	flag := ServeCmd.Flags().Lookup("metrics-port")
	if flag == nil || flag.DefValue != "9090" {
		t.Fatalf("metrics-port flag = %+v", flag)
	}
}

func TestServeUntilSignal(t *testing.T) {
	signals := make(chan os.Signal, 1)
	signals <- syscall.SIGTERM
	if err := serveUntilSignal(0, signals); err != nil {
		t.Fatalf("serveUntilSignal() error = %v", err)
	}
}
