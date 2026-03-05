package services

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestStepper_Markf(t *testing.T) {
	var buf bytes.Buffer
	logFn := func(format string, args ...interface{}) {
		buf.WriteString(strings.TrimSpace(fmt.Sprintf(format, args...)))
		buf.WriteByte('\n')
	}

	steps := []string{"start", "llm", "validation"}
	sp := NewStepper(steps, logFn)

	sp.Markf("start", "Processing %s", "ad1")
	sp.Markf("llm", "LLM done")
	sp.Markf("unknown", "Extra step")

	out := buf.String()
	if !strings.Contains(out, "[1/3 START]") {
		t.Fatalf("expected start step prefix, got: %s", out)
	}
	if !strings.Contains(out, "[2/3 LLM]") {
		t.Fatalf("expected llm step prefix, got: %s", out)
	}
	if !strings.Contains(out, "[4/3 UNKNOWN]") {
		t.Fatalf("expected unknown step to increment beyond total, got: %s", out)
	}
}
