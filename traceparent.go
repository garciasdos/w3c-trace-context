package tracecontext

import (
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

const FlagSampled uint8 = 1

var (
	traceIdFormat      = `[a-f0-9]{32}`
	traceIdPattern     = regexp.MustCompile(`^` + traceIdFormat + `$`)
	parentIdFormat     = `[a-f0-9]{16}`
	parentIdPattern    = regexp.MustCompile(`^` + parentIdFormat + `$`)
	traceParentPattern = regexp.MustCompile(`^[a-f0-9]{2}-` + traceIdFormat + `-` + parentIdFormat + `-[a-f0-9]{2}$`)
)

type TraceParent struct {
	Version  uint8
	TraceId  string
	ParentId string
	Flags    byte
}

func ParseTraceParent(s string) (*TraceParent, error) {
	parent := TraceParent{}

	if !traceParentPattern.MatchString(s) {
		return nil, errors.New("traceparent doesn't match the specified pattern")
	}

	parsedVersion, err := strconv.ParseInt(s[0:2], 16, 8)
	if err != nil {
		return nil, errors.New("cannot parse version")
	}

	parent.Version = uint8(parsedVersion)

	parent.TraceId = s[3:35]
	if parent.TraceId == "00000000000000000000000000000000" {
		return nil, errors.New("all zero trace id is not allowed")
	}

	parent.ParentId = s[36:52]
	if parent.ParentId == "0000000000000000" {
		return nil, errors.New("all zero trace id is not allowed")
	}

	parsedFlags, err := hex.DecodeString(s[53:55])
	if err != nil {
		return nil, errors.New("cannot parse flags")
	}
	parent.Flags = parsedFlags[0]

	return &parent, nil
}

func (p *TraceParent) IsSampled() bool {
	return p.Flags&FlagSampled != 0
}
func (tp *TraceParent) SetSampled(s bool) {
	if s {
		tp.Flags |= FlagSampled
	} else {
		tp.Flags &= ^FlagSampled
	}
}

func NewTraceParent(traceId string, parentId string) (*TraceParent, error) {
	if !traceIdPattern.MatchString(traceId) {
		return nil, errors.New("traceId doesn't match the specified pattern")
	}
	if !parentIdPattern.MatchString(parentId) {
		return nil, errors.New("parentId doesn't match the specified pattern")
	}

	tp := TraceParent{
		Version:  0,
		TraceId:  traceId,
		ParentId: parentId,
		Flags:    0,
	}

	return &tp, nil
}

func (tp *TraceParent) String() string {
	return fmt.Sprintf("%02x-%s-%s-%02x", tp.Version, tp.TraceId, tp.ParentId, tp.Flags)
}
