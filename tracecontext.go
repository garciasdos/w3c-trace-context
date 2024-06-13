package tracecontext

import (
	"errors"
	"fmt"
	"net/http"
)

const (
	TraceStateHeader  = "tracestate"
	TraceParentHeader = "traceparent"
)

// TraceContext combines the information of traceparent and tracestate.
// It can be parsed from HTTP headers and written to HTTP headers.
type TraceContext struct {
	TraceParent *TraceParent
	TraceState  *TraceState
}

// ParseTraceContext attempts to extract TraceContext information from a given
// set of headers. Partial data may be returned per the W3C specification.
// If parsing completely fails, an error is returned.
func ParseTraceContext(headers http.Header) (*TraceContext, error) {
	traceContext := TraceContext{}

	traceparentHeader := headers.Get(TraceParentHeader)
	traceParent, err := ParseTraceParent(traceparentHeader)
	// If the vendor failed to parse traceparent, it MUST NOT attempt to parse tracestate
	if err != nil {
		return nil, err
	}
	traceContext.TraceParent = traceParent

	tracestateHeader := headers.Get(TraceStateHeader)
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
// The TraceState will be mutated:
//   * parentId will be set to the provided value or generated randomly if left
//     empty
//   * The sampled flag will be set in accordance with the selected sampling
//     behavior
//   * member will be added to the tracestate if it is not nil
// The final mutated TraceContext based on which the headers were generated is
// returned as well.
func HandleTraceContext(headers *http.Header, parentId string, member *TraceStateMember, sampling SamplingBehavior) (*http.Header, *TraceContext, error) {
	newHeaders := headers.Clone()
	var newTraceContext *TraceContext

	if headers.Get(TraceParentHeader) != "" {
		tc, err := ParseTraceContext(*headers)
		if err != nil {
			// If parsing fails, the vendor creates a new traceparent header and
			// deletes the tracestate
			newHeaders.Del(TraceStateHeader)
			tc, err := GenerateTraceContext(parentId, member, sampling)
			if err != nil {
				return nil, nil, err
			}
			newTraceContext = tc
		} else {
			newTraceContext = tc
			newTraceContext.Mutate(parentId, sampling, member)
		}
	} else {
		// If a tracestate header is received without an accompanying
		// traceparent header, it is invalid and MUST be discarded.
		newHeaders.Del(TraceStateHeader)
		tc, err := GenerateTraceContext(parentId, member, sampling)
		if err != nil {
			return nil, nil, err
		}
		newTraceContext = tc
	}

	newTraceContext.WriteHeaders(&newHeaders)
	return &newHeaders, newTraceContext, nil
}

func HandleKongTraceContext(headers map[string][]string, parentId string, member *TraceStateMember, sampling SamplingBehavior) (*http.Header, *TraceContext, error) {
	httpHeaders := convertToHTTPHeader(headers)
	var newTraceContext *TraceContext

	if traceParent, exists := headers[TraceParentHeader]; exists && len(traceParent) > 0 {
		tc, err := ParseTraceContext(httpHeaders)
		if err != nil {
			// If parsing fails, the vendor creates a new traceparent header and
			// deletes the tracestate
			httpHeaders.Del(TraceStateHeader)
			tc, err := GenerateTraceContext(parentId, member, sampling)
			if err != nil {
				return nil, nil, err
			}
			newTraceContext = tc
		} else {
			newTraceContext = tc
			newTraceContext.Mutate(parentId, sampling, member)
		}
	} else {
		// If a tracestate header is received without an accompanying
		// traceparent header, it is invalid and MUST be discarded.

		httpHeaders.Del(TraceStateHeader)
		tc, err := GenerateTraceContext(parentId, member, sampling)
		if err != nil {
			return nil, nil, err
		}
		newTraceContext = tc
	}

	newTraceContext.WriteHeaders(&httpHeaders)

	return &httpHeaders, newTraceContext, nil
}

func convertToHTTPHeader(headers map[string][]string) http.Header {
	httpHeaders := http.Header{}
	for key, values := range headers {
		for _, value := range values {
			httpHeaders.Add(key, value)
		}
	}
	return httpHeaders
}

// GenerateTraceContext generates a new TraceContext object with a random
// trace id.
// The parentId will also be randomly generated if an empty one is provided.
// If the member is nil, no tracestate will be set
// If the member value is left empty, the generated
// parent id will be used as the vendor value.
// Errors will be returned if the random value generation fails or if the
// provided key or value don't match the allowed format.
func GenerateTraceContext(parentId string, member *TraceStateMember, sampling SamplingBehavior) (*TraceContext, error) {
	traceId, err := randomHex(16)
	if err != nil {
		return nil, err
	}
	if parentId == "" {
		parentId, err = randomHex(8)
		if err != nil {
			return nil, err
		}
	}

	tp, err := NewTraceParent(traceId, parentId)
	if err != nil {
		return nil, err
	}

	err = tp.applySamplingBehavior(sampling)
	if err != nil {
		return nil, err
	}

	var ts *TraceState
	if member == nil {
		ts = NewEmptyTraceState()
	} else {
		if member.Value == "" {
			member.Value = parentId
		}
		ts, err = NewTraceState(*member)
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

// NewTraceContext returns a new TraceContext object initialized with the
// provided traceId and parentId values.
// An error is returned if the provided values don't match the format
// specification.
func NewTraceContext(traceId string, parentId string) (*TraceContext, error) {
	tp, err := NewTraceParent(traceId, parentId)
	if err != nil {
		return nil, err
	}

	tc := TraceContext{
		TraceParent: tp,
		TraceState:  NewEmptyTraceState(),
	}

	return &tc, nil
}

// Mutate mutates the TraceContext object:
//   * The parentId is updated with the provided value. If an empty value is
//     provided, the parentId is randomly generated instead
//   * The sampled flag will be set in accordance with the selected sampling
//     behavior
//   * member is added to the tracestate list as long as member is not nil.
//     If member.Value is nil, the parentId is used as the value
func (tc *TraceContext) Mutate(parentId string, sampling SamplingBehavior, member *TraceStateMember) error {
	var err error
	if tc.TraceParent == nil {
		return errors.New("TraceContext without TraceParent cannot be mutated")
	}
	if parentId == "" {

		parentId, err = randomHex(8)
		if err != nil {
			return err
		}
	}
	err = tc.TraceParent.SetParentId(parentId)
	if err != nil {
		return err
	}

	err = tc.TraceParent.applySamplingBehavior(sampling)
	if err != nil {
		return err
	}

	if member != nil {
		if member.Value == "" {
			member.Value = parentId
		}
		if tc.TraceState == nil {
			tc.TraceState, err = NewTraceState(*member)
			if err != nil {
				return err
			}
		} else {
			tc.TraceState.Mutate(*member)
		}
	}
	return nil
}

// WriteHeaders writes the traceparent and tracestate headers to the provided
// headers object. Any existing headers of the same name are overwritten.
func (tc *TraceContext) WriteHeaders(headers *http.Header) {
	if tc.TraceParent != nil {
		headers.Set(TraceParentHeader, tc.TraceParent.String())
	}

	// Vendors MUST accept empty tracestate headers but SHOULD avoid sending them
	if tc.TraceState != nil && len(tc.TraceState.Members) > 0 {
		headers.Set(TraceStateHeader, tc.TraceState.String())
	}
}
