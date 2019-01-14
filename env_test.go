package main

import (
	"testing"
)

func TestEnv(t *testing.T) {
	e1 := NewEnv(nil, nil)
	a1, b1 := Number(1), Number(3)
	e1.Set("a", a1)
	e1.Set("b", b1)

	e2 := NewEnv(e1, nil)
	a2 := a1 + 3
	e2.Set("a", a2)

	if val := e1.Get("a"); val != a1 {
		t.Fatalf("bad e1['a']: %v", val)
	}

	if val := e2.Get("a"); val != a2 {
		t.Fatalf("bad e2['a']: %v", val)
	}

	if val := e2.Get("b"); val != b1 {
		t.Fatalf("bad e2['b']: %v", val)
	}

	if val := e2.Get("c"); val != nil {
		t.Fatalf("e2['c'] != nil (got %v)", val)
	}
}
