package tracecontext

import (
	"net/http"
	"testing"
)

func TestParseTraceContext(t *testing.T) {
	headers := http.Header{}
	headers.Add("traceparent", "00-0af7651916cd43dd8448eb211c80319c-00f067aa0ba902b7-01")
	headers.Add("tracestate", "vendor1=val1, vendor2=val2 ")

	tc, err := ParseTraceContext(headers)

	if err != nil {
		t.Error("Failed to parse trace context")
	}
	if tc == nil {
		t.Error("No trace context returned")
	}
	if tc != nil && tc.TraceParent == nil {
		t.Error("No trace parent returned")
	}
	if tc != nil && tc.TraceState == nil {
		t.Error("No trace state returned")
	}
}

func TestParseTraceContextEmptyTracestate(t *testing.T) {
	// Vendors MUST accept empty tracestate headers
	headers := http.Header{}
	headers.Add("traceparent", "00-0af7651916cd43dd8448eb211c80319c-00f067aa0ba902b7-01")
	headers.Add("tracestate", "")

	tc, err := ParseTraceContext(headers)

	if err != nil {
		t.Error("Failed to parse trace context")
	}
	if tc != nil {
		if tc.TraceState == nil {
			t.Error("No trace state returned")
		}
		if len(tc.TraceState.Members) != 0 {
			t.Error("TraceState list is not empty")
		}
	}
}

func TestGenerateTraceContext(t *testing.T) {
	vendorName := "vendor"
	vendorValue := "value"
	tc, err := GenerateTraceContext("", &TraceStateMember{Key: vendorName, Value: vendorValue}, SamplingBehaviorAlwaysSampled)

	if err != nil {
		t.Error("Failed to generate trace context")
	}
	if tc != nil {
		if tc.TraceState == nil {
			t.Error("No trace state generated")
		}
		if len(tc.TraceState.Members) != 1 {
			t.Error("TraceState list is not 1")
		}
		if tc.TraceState.Members[0].Key != vendorName {
			t.Error("Vendor name not set right")
		}
		if tc.TraceState.Members[0].Value != vendorValue {
			t.Error("Vendor value not set right")
		}
		if tc.TraceParent.IsSampled() != true {
			t.Error("Sampling flag not set right")
		}
	}
}

func TestHandleTraceContext(t *testing.T) {
	// Vendors MUST accept empty tracestate headers
	headers := http.Header{}
	headers.Add(TraceParentHeader, "00-0af7651916cd43dd8448eb211c80319c-00f067aa0ba902b7-01")
	headers.Add(TraceStateHeader, "vendor1=val1")

	newHeaders, tc, err := HandleTraceContext(&headers, "", &TraceStateMember{Key: "vendor2", Value: "val2"}, SamplingBehaviorNeverSampled)

	if err != nil {
		t.Error("Failed to handle trace context")
	}
	if tc != nil {
		if tc.TraceState == nil {
			t.Error("No trace state returned")
		}
		if len(tc.TraceState.Members) != 2 {
			t.Error("TraceState list is not as long as expected")
		}
		if tc.TraceParent.IsSampled() {
			t.Error("Trace not sampled")
		}
	}
	if newHeaders.Get(TraceStateHeader) != "vendor2=val2,vendor1=val1" {
		t.Error("TraceState is not as expected")
	}
	if newHeaders.Get(TraceParentHeader) == "" {
		t.Error("Missing traceparent header")
	}
}

func TestHandleTraceContextMissingHeaders(t *testing.T) {
	// Vendors MUST accept empty tracestate headers
	headers := http.Header{}

	newHeaders, tc, err := HandleTraceContext(&headers, "", &TraceStateMember{Key: "vendor2", Value: "val2"}, SamplingBehaviorAlwaysSampled)

	if err != nil {
		t.Error("Failed to handle trace context")
	}
	if tc != nil {
		if tc.TraceState == nil {
			t.Error("No trace state returned")
		}
		if len(tc.TraceState.Members) != 1 {
			t.Error("TraceState list is not as long as expected")
		}
		if !tc.TraceParent.IsSampled() {
			t.Error("Trace is not sampled")
		}
	}
	if newHeaders.Get(TraceStateHeader) != "vendor2=val2" {
		t.Error("TraceState is not as expected")
	}
	if newHeaders.Get(TraceParentHeader) == "" {
		t.Error("Missing traceparent header")
	}
}

func TestHandleTraceContextHigherVersion(t *testing.T) {
	// Vendors MUST accept empty tracestate headers
	headers := http.Header{}
	headers.Add(TraceParentHeader, "01-0af7651916cd43dd8448eb211c80319c-00f067aa0ba902b7-01")
	headers.Add(TraceStateHeader, "vendor1=val1")

	newHeaders, tc, err := HandleTraceContext(&headers, "", nil, SamplingBehaviorPassThrough)

	if err != nil {
		t.Error("Failed to handle trace context")
	}
	if tc != nil {
		if tc.TraceState == nil {
			t.Error("No trace state returned")
		}
		if len(tc.TraceState.Members) != 1 {
			t.Error("TraceState list is not as long as expected")
		}
		if !tc.TraceParent.IsSampled() {
			t.Error("Trace is not sampled")
		}
	}
	if newHeaders.Get(TraceStateHeader) != "vendor1=val1" {
		t.Error("TraceState is not as expected")
	}
	if newHeaders.Get(TraceParentHeader) == "" {
		t.Error("Missing traceparent header")
	}
}

func TestHandleTraceContextParsingError(t *testing.T) {
	// Vendors MUST accept empty tracestate headers
	headers := http.Header{}
	headers.Add(TraceParentHeader, "01-illegal")
	headers.Add(TraceStateHeader, "vendor1=val1")

	newHeaders, tc, err := HandleTraceContext(&headers, "", nil, SamplingBehaviorAlwaysSampled)

	if err != nil {
		t.Error("Failed to handle trace context")
	}
	if tc != nil {
		if !tc.TraceParent.IsSampled() {
			t.Error("Trace is not sampled")
		}
	}
	if newHeaders.Get(TraceStateHeader) != "" {
		t.Error("TraceState header returned")
	}
	if newHeaders.Get(TraceParentHeader) == "" {
		t.Error("Missing traceparent header")
	}
}

func TestHandleKongTraceContext(t *testing.T) {
	headers := map[string][]string{
		TraceParentHeader: {"00-0af7651916cd43dd8448eb211c80319c-00f067aa0ba902b7-01"},
		TraceStateHeader:  {"vendor1=val1"},
	}

	parentId := "00f067aa0ba902b7"
	member := &TraceStateMember{Key: "vendor2", Value: "val2"}
	sampling := SamplingBehaviorAlwaysSampled

	newHeaders, newTraceContext, err := HandleKongTraceContext(headers, parentId, member, sampling)
	if err != nil {
		t.Error("Failed to handle trace context:", err)
	}

	if newTraceContext == nil {
		t.Error("No trace context returned")
	} else {
		if !newTraceContext.TraceParent.IsSampled() {
			t.Error("Trace is not sampled")
		}
	}

	if newTraceState := newHeaders.Get(TraceStateHeader); newTraceState != "vendor2=val2,vendor1=val1" {
		t.Errorf("TraceState is not as expected: got %v", newTraceState)
	}

	if newTraceParent := newHeaders.Get(TraceParentHeader); newTraceParent != "00-0af7651916cd43dd8448eb211c80319c-00f067aa0ba902b7-01" {
		t.Errorf("Missing traceparent header: got %v", newTraceParent)
	}
}

func TestHandleKongTraceContextParsingError(t *testing.T) {
	headers := map[string][]string{
		TraceParentHeader: {"01-illegal"},
		TraceStateHeader:  {"vendor1=val1"},
	}
	parentId := "00f067aa0ba902b7"
	member := &TraceStateMember{Key: "vendor2", Value: "val2"}
	sampling := SamplingBehaviorAlwaysSampled

	newHeaders, newTraceContext, err := HandleKongTraceContext(headers, parentId, member, sampling)
	if err != nil {
		t.Error("Failed to handle trace context:", err)
	}

	if newTraceContext == nil {
		t.Error("No trace context returned")
	} else {
		if !newTraceContext.TraceParent.IsSampled() {
			t.Error("Trace is not sampled")
		}
	}

	if newTraceState := newHeaders.Get(TraceStateHeader); newTraceState != "vendor2=val2" {
		t.Errorf("TraceState header returned: got %v", newTraceState)
	}
	if newTraceParent := newHeaders.Get(TraceParentHeader); newTraceParent == "" {
		t.Errorf("Missing traceparent header: got %v", newTraceParent)
	}
}

func TestWriteHeadersEmpty(t *testing.T) {
	headers := http.Header{}
	tc := TraceContext{}
	tc.WriteHeaders(&headers)

	if headers.Get(TraceStateHeader) != "" {
		t.Error("tracestate header written")
	}
	if headers.Get(TraceParentHeader) != "" {
		t.Error("traceparent header written")
	}
}

func TestWriteHeadersNoEmptyTraceState(t *testing.T) {
	headers := http.Header{}
	ts := NewEmptyTraceState()
	tc := TraceContext{
		TraceState: ts,
	}
	tc.WriteHeaders(&headers)

	if headers.Get(TraceStateHeader) != "" {
		t.Error("tracestate header written")
	}
}

func TestNewTraceContext(t *testing.T) {
	traceId := "0af7651916cd43dd8448eb211c80319c"
	parentId := "00f067aa0ba902b7"
	tc, err := NewTraceContext(traceId, parentId)

	if err != nil {
		t.Error("Unexpected error: ", err)
	}

	if len(tc.TraceState.Members) != 0 {
		t.Error("Generated tracestate is not empty")
	}

	if tc.TraceParent.parentId != parentId || tc.TraceParent.traceId != traceId {
		t.Error("traceId or parentId not matching")
	}
}
