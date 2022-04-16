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
	tc, err := GenerateTraceContext(&TraceStateMember{Key: vendorName, Value: vendorValue})

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
	}
}

func TestHandleTraceContext(t *testing.T) {
	// Vendors MUST accept empty tracestate headers
	headers := http.Header{}
	headers.Add(TraceParentHeader, "00-0af7651916cd43dd8448eb211c80319c-00f067aa0ba902b7-00")
	headers.Add(TraceStateHeader, "vendor1=val1")

	newHeaders, tc, err := HandleTraceContext(&headers, &TraceStateMember{Key: "vendor2", Value: "val2"}, true)

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
		if !tc.TraceParent.IsSampled() {
			t.Error("Trace is not sampled")
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

	newHeaders, tc, err := HandleTraceContext(&headers, &TraceStateMember{Key: "vendor2", Value: "val2"}, true)

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

	newHeaders, tc, err := HandleTraceContext(&headers, nil, true)

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
