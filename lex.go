package main

import (
	"strings"
)

type Lexer struct {
	tokens []string
	i      int
}

func NewLexer(code string) *Lexer {
	code = strings.Replace(code, "(", " ( ", -1)
	code = strings.Replace(code, ")", " )", -1)
	tokens := strings.Fields(code)
	return &Lexer{tokens: tokens, i: -1}
}

func (l *Lexer) Next() bool {
	l.i++
	return l.i < len(l.tokens)
}

func (l *Lexer) Token() string {
	return l.tokens[l.i]
}

func (l *Lexer) Peek() string {
	if l.i >= len(l.tokens) {
		return ""
	}
	return l.tokens[l.i+1]
}
