package main

import (
	"fmt"
)

func evalDefine(args []Expression, env *Environment) (Object, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("wrong number of arguments for 'define'")
	}

	s, ok := args[0].(*SymbolExpr)
	if !ok {
		return nil, fmt.Errorf("bad name in 'define'")
	}

	val, err := args[1].Eval(env)
	if err != nil {
		return nil, err
	}
	env.Set(s.name, val)
	return val, nil
}

func evalSet(args []Expression, env *Environment) (Object, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("wrong number of arguments for 'define'")
	}

	s, ok := args[0].(*SymbolExpr)
	if !ok {
		return nil, fmt.Errorf("bad name in 'define'")
	}

	env = env.Find(s.name)
	if env == nil {
		return nil, fmt.Errorf("unknown name - %s", s.name)
	}

	val, err := args[1].Eval(env)
	if err != nil {
		return nil, err
	}

	env.Set(s.name, val)
	return val, nil
}

func evalIf(args []Expression, env *Environment) (Object, error) {
	if len(args) != 3 { // TODO: if without else
		return nil, fmt.Errorf("wrong number of arguments for 'define'")
	}

	cond, err := args[0].Eval(env)
	if err != nil {
		return nil, err
	}

	if cond == 1.0 {
		return args[1].Eval(env)
	}
	return args[2].Eval(env)
}

func evalOr(args []Expression, env *Environment) (Object, error) {
	for _, e := range args {
		obj, err := e.Eval(env)
		if err != nil {
			return nil, err
		}

		val, ok := obj.(float64)
		if !ok {
			return nil, fmt.Errorf("or - %v bad type %T", val, val)
		}

		if val != 0.0 {
			return val, nil
		}
	}

	return 0.0, nil
}

func evalAnd(args []Expression, env *Environment) (Object, error) {
	val, ok := 1.0, false
	for _, e := range args {
		obj, err := e.Eval(env)
		if err != nil {
			return nil, err
		}

		val, ok = obj.(float64)
		if !ok {
			return nil, fmt.Errorf("or - %v bad type %T", val, val)
		}

		if val == 0.0 {
			return val, nil
		}
	}

	return val, nil
}

func evalLambda(args []Expression, env *Environment) (Object, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("malformed lambda")
	}

	le, ok := args[0].(*ListExpr)
	if !ok {
		return nil, fmt.Errorf("malformed lambda")
	}

	params := make([]string, len(le.children))
	for i, e := range le.children {
		s, ok := e.(*SymbolExpr)
		if !ok {
			return nil, fmt.Errorf("malformed lambda")
		}
		params[i] = s.name
	}
	obj := &Lambda{
		env:    env,
		params: params,
		body:   args[1],
	}
	return obj, nil
}
