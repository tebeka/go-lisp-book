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
	env       = make(map[string]interface{})
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
		var children []*SExpr
		for tokens[0] != ")" {
			var child *SExpr
			child, tokens, err = readSExpr(tokens)
			if err != nil {
				return nil, nil, err
			}
			children = append(children, child)
		}
		sexpr.value = children[0]
		sexpr.children = children[1:]
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

func evalAtom(atom interface{}) (interface{}, error) {
	if val, ok := atom.(float64); ok {
		return val, nil
	}

	name, ok := atom.(string)
	if !ok {
		return nil, fmt.Errorf("unkown atom type for %v - %T", atom, atom)
	}

	value, ok
}

func eval(s *SExpr) (interface{}, error) {
	if len(s.children) == 0 { // atom
	}
	val, ok := s.value.(float64)
	if ok {
		if len(s.children) > 0 {
			return nil, fmt.Errorf("%s is not callable", s.value)
		}
		return val, nil
	}

	name, ok := s.value.(string)
	if !ok {
		return nil, fmt.Errorf("unknown type - %T", s.value)
	}

	val, ok := env[name]
	if !ok {
		return
	}

	if len(s.children) == 0 {
		if val, ok := s.value.(float64); ok {
			return val, nil
		}

		name, ok := s.value.(string)
		if !ok {
		}

		val, ok := env[name]
		if !ok {
			return nil, fmt.Errorf("unknown name - %q", name)
		}

		return val, nil
	}

	val, ok := env[name]
	if !ok {
		if !ok {
			return nil, fmt.Errorf("unknown name - %q", name)
		}
	}
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

// Env is environment
type Env struct {
	bindings map[string]interface{}
	parent   *Env
}

func printSExpr(s *SExpr, indent int) {
	fmt.Printf("%*s", indent, " ")
	fmt.Println(s.value)
	for _, c := range s.children {
		printSExpr(c, indent+4)
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
	printSExpr(sexpr, 0)
}
