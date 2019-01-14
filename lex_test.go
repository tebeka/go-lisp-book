package main

import (
	"testing"
)

func tokenize(code string) []string {
	var out []string
	lex := NewLexer(code)
	for lex.Next() {
		out = append(out, lex.Token())
	}

	return out
}

func sliceEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i, v := range s1 {
		if s2[i] != v {
			return false
		}
	}

	return true
}

func TestLexer(t *testing.T) {
	code := "(+ 3 4)"
	expected := []string{"(", "+", "3", "4", ")"}

	tokens := tokenize(code)
	if !sliceEqual(tokens, expected) {
		t.Fatalf("bad tokens: %v != %v", tokens, expected)
	}
}

type testCase struct {
	code   string
	tokens []string
}

var testCases = []testCase{
	{"1", []string{"1"}},
	{"(+ 3 4)", []string{"(", "+", "3", "4", ")"}},
}

func TestLexerMany(t *testing.T) {
	for _, tc := range testCases {
		t.Run(tc.code, func(t *testing.T) {
			tokens := tokenize(tc.code)
			if !sliceEqual(tokens, tc.tokens) {
				t.Fatalf("%v != %v", tokens, tc.tokens)
			}
		})
	}
}
