package tracecontext

import (
	"testing"
)

func TestParseTraceState(t *testing.T) {
	_, err := ParseTraceState("congo=congosSecondPosition,  rojo=rojosFirstPosition  ")

	if err != nil {
		t.Error("Failed to parse traceparent")
	}
}

func TestMutate(t *testing.T) {
	member1 := TraceStateMember{
		Key:   "member1",
		Value: "value1",
	}
	member2 := TraceStateMember{
		Key:   "member2",
		Value: "value2",
	}
	ts := TraceState{
		Members: []*TraceStateMember{
			&member1,
			&member2,
		},
	}

	member3Key := "member3"
	member3Val := "value3"
	ts.Mutate(member3Key, member3Val)

	if len(ts.Members) != 3 {
		t.Errorf("Incorrect length %d after mutate", len(ts.Members))
	}
	if ts.Members[0].Value != member3Val || ts.Members[0].Key != member3Key {
		t.Error("Key or value of new member are not correct")
	}
}

func TestMutateReplace(t *testing.T) {
	member1 := TraceStateMember{
		Key:   "member1",
		Value: "value1",
	}
	member2 := TraceStateMember{
		Key:   "member2",
		Value: "value2",
	}
	ts := TraceState{
		Members: []*TraceStateMember{
			&member1,
			&member2,
		},
	}

	member2Key := "member2"
	member2NewVal := "newVal"
	ts.Mutate(member2Key, member2NewVal)

	if len(ts.Members) != 2 {
		t.Errorf("Incorrect length %d after mutate", len(ts.Members))
	}
	if ts.Members[0].Value != member2NewVal || ts.Members[0].Key != member2Key {
		t.Error("Key or value of new member are not correct")
	}
}

func TestMutateIllegalKey(t *testing.T) {
	member1 := TraceStateMember{
		Key:   "member1",
		Value: "value1",
	}
	member2 := TraceStateMember{
		Key:   "member2",
		Value: "value2",
	}
	ts := TraceState{
		Members: []*TraceStateMember{
			&member1,
			&member2,
		},
	}

	err := ts.Mutate("IllegalKey", "val")
	if err == nil {
		t.Error("Illegal key didn't cause an error")
	}
}

func TestMutateIllegalValue(t *testing.T) {
	member1 := TraceStateMember{
		Key:   "member1",
		Value: "value1",
	}
	member2 := TraceStateMember{
		Key:   "member2",
		Value: "value2",
	}
	ts := TraceState{
		Members: []*TraceStateMember{
			&member1,
			&member2,
		},
	}

	err := ts.Mutate("key", "\nval")
	if err == nil {
		t.Error("Illegal value didn't cause an error")
	}
}

func TestString(t *testing.T) {
	member1 := TraceStateMember{
		Key:   "member1",
		Value: "value1",
	}
	member2 := TraceStateMember{
		Key:   "member2",
		Value: "value2",
	}
	ts := TraceState{
		Members: []*TraceStateMember{
			&member1,
			&member2,
		},
	}

	s := ts.String()
	if s != "member1=value1,member2=value2" {
		t.Errorf("Wrong string value returned: '%s'", s)
	}
}
