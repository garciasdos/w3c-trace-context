package tracecontext

import (
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
)

type SamplingBehavior uint8

const (
	FlagSampled                         uint8 = 1
	HighestSupportedTraceContextVersion uint8 = 0

	// SamplingBehaviorPassThrough leads to sampling decisions from the calling
	// system to be passed through. If new traces are generated, their sampling
	// flag will not be set.
	SamplingBehaviorPassThrough SamplingBehavior = 0
	// SamplingBehaviorAlwaysSampled always overrides the sampling flag to true
	SamplingBehaviorAlwaysSampled SamplingBehavior = 1
	// SamplingBehaviorNeverSampled always overrides the sampling flag to false
	SamplingBehaviorNeverSampled SamplingBehavior = 2
)

var (
	traceIdFormat          = `[a-f0-9]{32}`
	traceIdPattern         = regexp.MustCompile(`^` + traceIdFormat + `$`)
	traceIdAndDashPattern  = regexp.MustCompile(traceIdFormat + `-`)
	parentIdFormat         = `[a-f0-9]{16}`
	parentIdPattern        = regexp.MustCompile(`^` + parentIdFormat + `$`)
	parentIdAndDashPattern = regexp.MustCompile(parentIdFormat + `-`)
	versionFormat          = `^[a-f0-9]{2}-`
	versionPattern         = regexp.MustCompile(versionFormat)
	flagsFormat            = `[a-f0-9]{2}`
	flagsPattern           = regexp.MustCompile(flagsFormat)
	traceParentPattern     = regexp.MustCompile(
		versionFormat + traceIdFormat + `-` + parentIdFormat + `-` + flagsFormat + `$`)
)

// TraceParent represents the information contained in the traceparent header
type TraceParent struct {
	version  uint8
	traceId  string
	parentId string
	flags    byte
}

// ParseTraceParent parses the input string and - on success - returns a
// TraceParent object
func ParseTraceParent(s string) (*TraceParent, error) {
	parent := TraceParent{}

	if !traceParentPattern.MatchString(s) {
		// When the version prefix cannot be parsed (it's not 2 hex characters
		// followed by a dash (-)), the implementation should restart the trace.
		if !versionPattern.MatchString(s) {
			return nil, errors.New("cannot parse traceparent version")
		}

		versionByte, err := hex.DecodeString(s[0:2])
		if err != nil {
			return nil, errors.New("cannot parse traceparent version")
		}
		parsedVersion := uint8(versionByte[0])

		// If a higher version is detected, the implementation SHOULD try to
		// parse it by trying the following
		if parsedVersion > HighestSupportedTraceContextVersion {
			return parseHigherVersion(s)
		}

		return nil, errors.New("traceparent doesn't match the specified pattern")
	}

	versionByte, err := hex.DecodeString(s[0:2])
	parsedVersion := uint8(versionByte[0])
	if err != nil {
		return nil, errors.New("cannot parse version")
	}
	// Version ff is invalid
	if parsedVersion == 255 {
		return nil, errors.New("version 'ff' is invalid")
	}

	parent.version = uint8(parsedVersion)

	parent.traceId = s[3:35]
	if parent.traceId == "00000000000000000000000000000000" {
		return nil, errors.New("all zero trace id is not allowed")
	}

	parent.parentId = s[36:52]
	if parent.parentId == "0000000000000000" {
		return nil, errors.New("all zero parent id is not allowed")
	}

	parsedFlags, err := hex.DecodeString(s[53:55])
	if err != nil {
		return nil, errors.New("cannot parse flags")
	}
	parent.flags = parsedFlags[0]

	return &parent, nil
}

// parseHigherVersion contains the logic to attempt to parse a traceparent that
// has a version higher than 00.
func parseHigherVersion(s string) (*TraceParent, error) {
	// If the size of the header is shorter than 55 characters, the
	// vendor should not parse the header and should restart the trace.
	if len(s) < 55 {
		return nil, errors.New("traceparent is shorter than 55 characters")
	}

	// Parse trace-id (from the first dash through the next 32 characters).
	// Vendors MUST check that the 32 characters are hex, and that they are
	// followed by a dash (-)
	if !traceIdAndDashPattern.MatchString(s[3:37]) {
		return nil, errors.New("cannot parse trace id")
	}
	traceId := s[3:35]

	// Parse parent-id (from the second dash at the 35th position through the
	// next 16 characters). Vendors MUST check that the 16 characters are hex
	// and followed by a dash.
	if !parentIdAndDashPattern.MatchString(s[36:53]) {
		return nil, errors.New("cannot parse parent id")
	}
	parentId := s[36:52]

	// Parse the sampled bit of flags (2 characters from the third dash).
	if !flagsPattern.MatchString(s[53:55]) {
		return nil, errors.New("cannot parse flags")
	}
	// Vendors MUST check that the 2 characters are either the end of the
	// string or a dash.
	if !(len(s) == 55 || (len(s) >= 56 && s[55] == '-')) {
		return nil, errors.New("flags not followed by end of string or dash")
	}

	parsedFlags, err := hex.DecodeString(s[53:55])
	if err != nil {
		return nil, errors.New("cannot parse flags")
	}
	flags := parsedFlags[0]

	// Vendors MUST use these fields to construct the new traceparent field
	// according to the highest version of the specification known to the
	// implementation (in this specification it is 00).
	tp := TraceParent{
		version:  HighestSupportedTraceContextVersion,
		traceId:  traceId,
		parentId: parentId,
		flags:    flags,
	}

	return &tp, nil
}

// IsSampled returns true if the sampled flag in the TraceParent is set
func (p *TraceParent) IsSampled() bool {
	return p.flags&FlagSampled != 0
}

// SetSampled updates the sampled flag with the given value
func (tp *TraceParent) SetSampled(s bool) {
	if s {
		tp.flags |= FlagSampled
	} else {
		tp.flags &= ^FlagSampled
	}
}

// NewTraceParent generates a new TraceParent based on the provided values.
// If the values don't match the correct format, an error is returned
func NewTraceParent(traceId string, parentId string) (*TraceParent, error) {
	tp := TraceParent{}
	err := tp.SetTraceId(traceId)
	if err != nil {
		return nil, err
	}
	err = tp.SetParentId(parentId)
	if err != nil {
		return nil, err
	}

	return &tp, nil
}

func (tp *TraceParent) ParentId() string {
	return tp.parentId
}

func (tp *TraceParent) TraceId() string {
	return tp.traceId
}

func (tp *TraceParent) Version() uint8 {
	return tp.version
}

func (tp *TraceParent) SetParentId(parentId string) error {
	if !parentIdPattern.MatchString(parentId) {
		return errors.New("parentId doesn't match the specified pattern")
	}
	tp.parentId = parentId
	return nil
}

func (tp *TraceParent) SetTraceId(traceId string) error {
	if !traceIdPattern.MatchString(traceId) {
		return errors.New("traceId doesn't match the specified pattern")
	}
	tp.traceId = traceId
	return nil
}

// String returns the string representation of the TraceParent
func (tp *TraceParent) String() string {
	return fmt.Sprintf("%02x-%s-%s-%02x",
		tp.version,
		tp.traceId,
		tp.parentId,
		tp.flags)
}

// applySamplingBehavior applies the selected sampling behavior to the TraceParent
func (tp *TraceParent) applySamplingBehavior(sampling SamplingBehavior) error {
	switch sampling {
	case SamplingBehaviorPassThrough:
		// Nothing to do to retain the previous value
	case SamplingBehaviorAlwaysSampled:
		tp.SetSampled(true)
	case SamplingBehaviorNeverSampled:
		tp.SetSampled(false)
	default:
		return errors.New("invalid sampling behavior")
	}
	return nil
}
