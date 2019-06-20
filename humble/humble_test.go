package main

import (
	"io/ioutil"
	"testing"
)

func run(t *testing.T, code string) Object {
	tokens := Tokenize(code)
	expr, _, err := ReadExpr(tokens)
	if err != nil {
		t.Fatal("read expression")
	}

	obj, err := expr.Eval(builtins)
	if err != nil {
		t.Fatal("eval")
	}

	return obj
}

var evalTestCases = []struct {
	fileName string
	expr     string
	out      Object
}{
	{"fact.scm", "(fact 10)", 3628800.0},
	{"collatz.scm", "(collatz 7)", 22.0},
}

func TestEval(t *testing.T) {
	for _, tc := range evalTestCases {
		t.Run(tc.fileName, func(t *testing.T) {
			data, err := ioutil.ReadFile(tc.fileName)
			if err != nil {
				t.Fatal("open")
			}
			run(t, string(data))

			out := run(t, tc.expr)
			if tc.out != out {
				t.Fatalf("result mismatch: %#v != %#v", tc.out, out)
			}
		})
	}
}

var logicTestCases = []struct {
	expr string
	out  Object
}{
	{"(or)", 0.0},
	{"(or 1 2)", 1.0},
	{"(or 0 2 1)", 2.0},
	{"(and)", 1.0},
	{"(and 1 2)", 2.0},
	{"(and 1 0 3)", 0.0},
}

func TestLogic(t *testing.T) {
	for _, tc := range logicTestCases {
		t.Run(tc.expr, func(t *testing.T) {
			out := run(t, tc.expr)
			if tc.out != out {
				t.Fatalf("result mismatch: %#v != %#v", tc.out, out)
			}
		})
	}
}
