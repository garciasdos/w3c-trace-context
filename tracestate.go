package tracecontext

import (
	"errors"
	"regexp"
	"strings"
)

var (
	keyFormat     = `[a-z0-9][a-z0-9_\-\*\/@]{0,255}`
	keyPattern    = regexp.MustCompile(`^` + keyFormat + `$`)
	valueFormat   = `[\x20-\x2b\x2d-\x3c\x3e-\x7e]{0,255}[\x21-\x2b\x2d-\x3c\x3e-\x7e]`
	valuePattern  = regexp.MustCompile(`^` + valueFormat + `$`)
	memberFormat  = `\s*(` + keyFormat + `)=(` + valueFormat + `)\s*`
	memberPattern = regexp.MustCompile(`^` + memberFormat + `$`)
)

// TraceState represents the information contained in the tracestate header
type TraceState struct {
	Members []*TraceStateMember
}

// TraceStateMember represents a single entry in the TraceState list
type TraceStateMember struct {
	Key   string
	Value string
}

// ParseTraceState parses the provided string and - on success - returns a
// TraceState object
func ParseTraceState(s string) (*TraceState, error) {
	candidates := strings.Split(s, ",")

	traceState := TraceState{}
	for _, candidate := range candidates {
		if len(candidate) == 0 {
			continue
		}
		member, err := parseMember(candidate)
		if err != nil {
			return nil, err
		}
		traceState.Members = append(traceState.Members, member)
	}

	return &traceState, nil
}

func parseMember(s string) (*TraceStateMember, error) {
	matches := memberPattern.FindStringSubmatch(s)
	if len(matches) != 3 {
		return nil, errors.New("invalid number of matches")
	}

	member := TraceStateMember{
		Key:   matches[1],
		Value: matches[2],
	}

	return &member, nil
}

// Mutate will add a new member to beginning of the list and - if the key is
// already present - remove the old entry
func (ts *TraceState) Mutate(member TraceStateMember) error {
	if !keyPattern.MatchString(member.Key) {
		return errors.New("key doesn't match allowed key pattern")
	}
	if !valuePattern.MatchString(member.Value) {
		return errors.New("value doesn't match allowed value pattern")
	}

	idx := -1
	for i := range ts.Members {
		if ts.Members[i].Key == member.Key {
			idx = i
			break
		}
	}

	// If the member already exists in the list, the old entry needs to be
	// removed first
	if idx != -1 {
		if idx == len(ts.Members)-1 { // If it's the last, it can easily be removed
			ts.Members = ts.Members[:idx]
		} else {
			copy(ts.Members[idx:], ts.Members[idx+1:])
			ts.Members = ts.Members[:len(ts.Members)-1]
		}
	}

	// Modified keys MUST be moved to the beginning (left) of the list
	ts.Members = append([]*TraceStateMember{&member}, ts.Members...)

	// If adding an entry would cause the tracestate list to contain more than
	// 32 list-members the right-most list-member should be removed from the list
	if len(ts.Members) > 32 {
		ts.Members = ts.Members[:32]
	}
	return nil
}

// String returns the string representation of the tracestate header value
func (ts *TraceState) String() string {
	sb := strings.Builder{}

	for i, m := range ts.Members {
		sb.WriteString(m.Key)
		sb.WriteString("=")
		sb.WriteString(m.Value)
		if i < len(ts.Members)-1 {
			sb.WriteString(",")
		}
	}

	return sb.String()
}

// MemberValue returns the string value of the member with the provided key.
// If the member doesn't exist, an empty string is returned.
func (ts *TraceState) MemberValue(memberKey string) string {
	for _, m := range ts.Members {
		if m.Key == memberKey {
			return m.Value
		}
	}
	return ""
}

// NewEmptyTraceState generates an empty TraceState object
func NewEmptyTraceState() *TraceState {
	ts := TraceState{}
	return &ts
}

// NewTraceState generates a TraceState object and adds an entry based on
// key and value
func NewTraceState(member TraceStateMember) (*TraceState, error) {
	ts := TraceState{}
	err := ts.Mutate(member)
	if err != nil {
		return nil, err
	}
	return &ts, nil
}
