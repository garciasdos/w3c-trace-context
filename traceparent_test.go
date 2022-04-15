package tracecontext

import (
	"testing"
)

func TestParseTraceParent(t *testing.T) {
	_, err := ParseTraceParent("00-0af7651916cd43dd8448eb211c80319c-00f067aa0ba902b7-01")

	if err != nil {
		t.Error("Failed to parse traceparent")
	}
}

func TestParseTraceParentZeroTraceId(t *testing.T) {
	_, err := ParseTraceParent("00-00000000000000000000000000000000-0000000000000001-00")

	if err == nil {
		t.Error("Parsed invalid traceparent")
	}
}

func TestParseTraceParentZeroTraceParent(t *testing.T) {
	_, err := ParseTraceParent("00-00000000000000000000000000000001-0000000000000000-00")

	if err == nil {
		t.Error("Parsed invalid traceparent")
	}
}

func TestParseTraceParentNonZeroVersion(t *testing.T) {
	_, err := ParseTraceParent("01-00000000000000000000000000000000-0000000000000000-00")

	if err == nil {
		t.Error("Parsed invalid version")
	}
}

func TestTraceParentIsSampled(t *testing.T) {
	sampled, _ := ParseTraceParent("00-0af7651916cd43dd8448eb211c80319c-00f067aa0ba902b7-01")
	notSampled, _ := ParseTraceParent("00-0af7651916cd43dd8448eb211c80319c-00f067aa0ba902b7-00")

	if !sampled.IsSampled() {
		t.Error("Sampled not detected even thought it was sampled")
	}

	if notSampled.IsSampled() {
		t.Error("Sampled detected even thought it wasn't sampled")
	}

}
