package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestContextLoggerHelpers(t *testing.T) {
	logger := &Logger{level: LevelDebug, format: FormatJSON, writer: &bytes.Buffer{}, noColor: true, fields: map[string]interface{}{}}
	defaultLogger = logger
	with := WithFields(map[string]interface{}{"request": "r1"})
	if with.fields["request"] != "r1" {
		t.Fatalf("fields not retained: %#v", with.fields)
	}
	if WithVDB("weaviate", "query").fields["vdb_type"] != "weaviate" {
		t.Fatal("WithVDB missing field")
	}
	if WithCollection("weaviate", "Docs").fields["collection"] != "Docs" {
		t.Fatal("WithCollection missing field")
	}
	if WithDocument("weaviate", "Docs", "doc-1").fields["document_id"] != "doc-1" {
		t.Fatal("WithDocument missing field")
	}
	logger.Info("hello %s", "world")
	logger.Debug("debug")
	logger.Warn("warn")
	logger.Error("error")
	if !strings.Contains(logger.writer.(*bytes.Buffer).String(), "hello") {
		t.Fatal("expected log output")
	}
}
