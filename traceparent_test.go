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

func TestParseTraceParentFFVersion(t *testing.T) {
	_, err := ParseTraceParent("ff-0af7651916cd43dd8448eb211c80319c-00f067aa0ba902b7-01")

	if err == nil {
		t.Error("Incorrectly parsed ff version")
	}
}

func TestParseTraceParentVersionOne(t *testing.T) {
	tp, err := ParseTraceParent("01-0af7651916cd43dd8448eb211c80319c-00f067aa0ba902b7-01-123")

	if err != nil {
		t.Error("Could not parse valid future version:", err)
	}
	if tp.version != HighestSupportedTraceContextVersion {
		t.Error("version not parsed correctly")
	}
	if tp.traceId != "0af7651916cd43dd8448eb211c80319c" {
		t.Error("trace id not parsed correctly", tp.traceId)
	}
	if tp.parentId != "00f067aa0ba902b7" {
		t.Error("parent id not parsed correctly")
	}
	if !tp.IsSampled() {
		t.Error("sampled flag parsed correctly")
	}

	// Verify that version gets downgraded
	s := tp.String()
	newTp, err := ParseTraceParent(s)
	if err != nil {
		t.Error("Could not parse generated trace parent:", err)
	}
	if newTp.Version() != HighestSupportedTraceContextVersion {
		t.Error("trace context version wasn't downgraded")
	}
}

func TestParseTraceParentFutureVersionNotFollowingPattern(t *testing.T) {
	_, err := ParseTraceParent("01-0af7651916cd43dd8448eb211c80319c1-00f067aa0ba902b7-01")

	if err == nil {
		t.Error("Incorrectly parsed invalid trace parent of future version")
	}
}

func TestParseTraceParentFutureVersionTooLongFlags(t *testing.T) {
	_, err := ParseTraceParent("01-0af7651916cd43dd8448eb211c80319c-00f067aa0ba902b7-010")

	if err == nil {
		t.Error("Incorrectly parsed invalid trace parent of future version")
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
