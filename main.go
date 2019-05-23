package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	commentRe = regexp.MustCompile(";.*$")
)

func tokenize(code string) []string {
	code = commentRe.ReplaceAllString(code, "")
	code = strings.Replace(code, "(", " ( ", -1)
	code = strings.Replace(code, ")", " )", -1)
	return strings.Fields(code)
}

// SExpr is a node in evaluation tree
type SExpr struct {
	value    interface{}
	children []*SExpr
}

func readSExpr(tokens []string) (*SExpr, []string, error) {
	var err error
	if len(tokens) == 0 {
		return nil, nil, io.EOF
	}
	tok, tokens := tokens[0], tokens[1:]
	var sexpr SExpr
	if tok == "(" {
		for tokens[0] != ")" {
			var child *SExpr
			child, tokens, err = readSExpr(tokens)
			if err != nil {
				return nil, nil, err
			}
			sexpr.children = append(sexpr.children, child)
		}
		tokens = tokens[1:] // remove closing ')'
		return &sexpr, tokens, nil
	}

	// TODO: file:line
	if tok == ")" {
		return nil, nil, fmt.Errorf("unexpected ')'")
	}

	/*
	   if tok in {'#t', '#f'}:
	       return tok == '#t'
	*/

	val, err := strconv.ParseFloat(tok, 64)
	if err == nil {
		sexpr.value = val
	} else {
		sexpr.value = tok
	}

	return &sexpr, tokens, nil
}

func repl() {
	fmt.Printf("» ")
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		tokens := tokenize(s.Text())
		fmt.Println(tokens)
		readSExpr(tokens)
		fmt.Printf("» ")
		/*
			code, err := parse(s.Text())
			if err != nil {
				fmt.Printf("syntax error: %s\n", err)
			}

			val, err := code.Eval()
			if err != nil {
				fmt.Printf("eval error: %s\n", err)
			}
			fmt.Println(val)
			fmt.Printf(">>> ")
		*/
	}

	switch s.Err() {
	case io.EOF, nil:
		// OK
	default:
		fmt.Printf("input error: %s\n", s.Err())
	}
}

func main() {
	code := "(* 3 6)"
	tokens := tokenize(code)
	sexpr, _, err := readSExpr(tokens)
	if err != nil {
		fmt.Printf("ERROR: %s", err)
		os.Exit(1)
	}
	fmt.Println(sexpr)
}

type Env struct {
	bindings map[string]Expr
	parent   *Env
}

func NewEnv(parent *Env, bindings map[string]Expr) *Env {
	if bindings == nil {
		bindings = make(map[string]Expr)
	}
	return &Env{bindings: bindings, parent: parent}
}

func (e *Env) Set(name string, value Expr) {
	e.bindings[name] = value
}

func (e *Env) Get(name string) Expr {
	for e != nil {
		val, ok := e.bindings[name]
		if ok {
			return val
		}
		e = e.parent
	}

	return nil
}
