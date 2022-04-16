package tracecontext

import (
	"fmt"
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

	member3 := TraceStateMember{Key: "member3", Value: "value3"}
	ts.Mutate(member3)

	if len(ts.Members) != 3 {
		t.Errorf("Incorrect length %d after mutate", len(ts.Members))
	}
	if ts.Members[0].Value != member3.Value || ts.Members[0].Key != member3.Key {
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

	member2new := TraceStateMember{Key: "member2", Value: "newVal"}
	ts.Mutate(member2new)

	if len(ts.Members) != 2 {
		t.Errorf("Incorrect length %d after mutate", len(ts.Members))
	}
	if ts.Members[0].Value != member2new.Value || ts.Members[0].Key != member2new.Key {
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

	err := ts.Mutate(TraceStateMember{Key: "IllegalKey", Value: "val"})
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

	err := ts.Mutate(TraceStateMember{Key: "key", Value: "\nval"})
	if err == nil {
		t.Error("Illegal value didn't cause an error")
	}
}

func TestMutateMaximumValues(t *testing.T) {
	ts := TraceState{
		Members: []*TraceStateMember{},
	}

	for i := 0; i < 32; i++ {
		m := TraceStateMember{
			Key:   fmt.Sprintf("m%d", i),
			Value: "v",
		}
		ts.Members = append(ts.Members, &m)
	}

	member3 := TraceStateMember{Key: "member3", Value: "value3"}
	ts.Mutate(member3)

	if len(ts.Members) != 32 {
		t.Errorf("Incorrect length %d after mutate", len(ts.Members))
	}
	if ts.Members[0].Value != member3.Value || ts.Members[0].Key != member3.Key {
		t.Error("Key or value of new member are not correct")
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
