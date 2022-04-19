package tracecontext

import (
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
)

const FlagSampled uint8 = 1

var (
	traceIdFormat      = `[a-f0-9]{32}`
	traceIdPattern     = regexp.MustCompile(`^` + traceIdFormat + `$`)
	parentIdFormat     = `[a-f0-9]{16}`
	parentIdPattern    = regexp.MustCompile(`^` + parentIdFormat + `$`)
	traceParentPattern = regexp.MustCompile(`^[a-f0-9]{2}-` + traceIdFormat + `-` + parentIdFormat + `-[a-f0-9]{2}$`)
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
	return fmt.Sprintf("%02x-%s-%s-%02x", tp.version, tp.traceId, tp.parentId, tp.flags)
}
