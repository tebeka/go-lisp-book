package main

import (
	"testing"
)

func TestParse(t *testing.T) {
	t.Log("TODO")
	code := "(+ 1 3)"
	expr, err := parse(code)
	if err != nil {
		t.Fatal(err)
	}

	val, err := expr.Eval()
	if err != nil {
		t.Fatal(err)
	}

	expected := Number(4)
	if val != expected {
		t.Fatalf("%q -> %v (expected %v)", code, val, expected)
	}

}
