package tracecontext

import (
	"net/http"
)

const (
	TRACESTATE_HEADER  = "tracestate"
	TRACEPARENT_HEADER = "traceparent"
)

type TraceContext struct {
	TraceParent *TraceParent
	TraceState  *TraceState
}

// ParseTraceContext attempts to extract TraceContext information from a given
// set of headers. Partial data may be returned per the W3C specification.
// If parsing completely fails, an error is returned.
func ParseTraceContext(headers http.Header) (*TraceContext, error) {
	traceContext := TraceContext{}

	traceparentHeader := headers.Get(TRACEPARENT_HEADER)
	traceParent, err := ParseTraceParent(traceparentHeader)
	// If the vendor failed to parse traceparent, it MUST NOT attempt to parse tracestate
	if err != nil {
		return nil, err
	}
	traceContext.TraceParent = traceParent

	tracestateHeader := headers.Get(TRACESTATE_HEADER)
	traceState, err := ParseTraceState(tracestateHeader)
	//failure to parse tracestate MUST NOT affect the parsing of traceparent
	if err == nil {
		traceContext.TraceState = traceState
	}

	return &traceContext, nil
}

// HandleTraceContext implements handling of the trace context read from the
// input headers. A copy of the headers is returned with only traceparent and
// tracestate modified.
// If no trace context information is present, it will be added.
// The final TraceContext based on which the headers were generated is returned
// as well.
// if vendorKey is not empty, the tracestate will be modified accordingly
func HandleTraceContext(headers *http.Header, vendorKey string, vendorValue string, sampled bool) (*http.Header, *TraceContext, error) {
	newHeaders := headers.Clone()
	var newTraceContext *TraceContext

	if headers.Get(TRACEPARENT_HEADER) != "" {
		tc, err := ParseTraceContext(*headers)
		if err != nil {
			// If parsing fails, the vendor creates a new traceparent header and
			// deletes the tracestate
			newHeaders.Del(TRACESTATE_HEADER)
			tc, err := GenerateTraceContext(vendorKey, vendorValue)
			if err != nil {
				return nil, nil, err
			}
			newTraceContext = tc
		} else {
			if vendorKey != "" {
				tc.TraceState.Mutate(vendorKey, vendorValue)
			}
			newTraceContext = tc
		}
	} else {
		// If a tracestate header is received without an accompanying
		// traceparent header, it is invalid and MUST be discarded.
		newHeaders.Del(TRACESTATE_HEADER)
		tc, err := GenerateTraceContext(vendorKey, vendorValue)
		if err != nil {
			return nil, nil, err
		}
		newTraceContext = tc
	}

	newTraceContext.TraceParent.SetSampled(sampled)
	newTraceContext.WriteHeaders(&newHeaders)
	return &newHeaders, newTraceContext, nil
}

// GenerateTraceContext generates a new TraceContext object with a random
// trace id and parent id.
// If the vendorKey is left empty, no tracestate will be set
// If the vendorValue is left empty, the generated
// parent id will be used as the vendor value.
// Errors will be returned if the random value generation fails or if the
// provided key or value don't match the allowed format.
func GenerateTraceContext(vendorKey string, vendorValue string) (*TraceContext, error) {
	traceId, err := randomHex(16)
	if err != nil {
		return nil, err
	}
	parentId, err := randomHex(8)
	if err != nil {
		return nil, err
	}

	tp, err := NewTraceParent(traceId, parentId)
	if err != nil {
		return nil, err
	}

	if vendorValue == "" {
		vendorValue = parentId
	}
	var ts *TraceState
	if vendorKey == "" {
		ts = NewEmptyTraceState()
	} else {
		ts, err = NewTraceState(vendorKey, vendorValue)
		if err != nil {
			return nil, err
		}
	}

	tc := TraceContext{
		TraceParent: tp,
		TraceState:  ts,
	}

	return &tc, nil
}

// WriteHeaders writes the traceparent and tracestate headers to the provided
// headers object. Any existing headers of the same name are overwritten.
func (tc *TraceContext) WriteHeaders(headers *http.Header) {
	headers.Set(TRACEPARENT_HEADER, tc.TraceParent.String())
	headers.Set(TRACESTATE_HEADER, tc.TraceState.String())
}
