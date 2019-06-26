package main

import (
	"fmt"
)

// Function object
type Function func(args []Object) (Object, error)

// Call implements Callable
func (f Function) Call(args []Object) (Object, error) {
	return f(args)
}

// Plus function. e.g. (+ 1 2 3)
func Plus(args []Object) (Object, error) {
	total := 0.0
	for i, arg := range args {
		fval, ok := arg.(float64)
		if !ok {
			return nil, fmt.Errorf("%d bad argument: %v of %T", i, args, arg)
		}
		total += fval
	}

	return total, nil
}

// Mul function. e.g. (* 1 2 3)
func Mul(args []Object) (Object, error) {
	total := 1.0
	for i, arg := range args {
		fval, ok := arg.(float64)
		if !ok {
			return nil, fmt.Errorf("%d bad argument: %v of %T", i, args, arg)
		}
		total *= fval
	}

	return total, nil
}

// Begin function. e.g. (begin (* 3 4) (/ 5 7))
func Begin(args []Object) (Object, error) {
	if len(args) == 0 {
		return 0.0, nil
	}

	return args[len(args)-1], nil
}

// BinOp is a binary (two argument) operator/function
type BinOp struct {
	name string
	op   func(float64, float64) (Object, error)
}

// Call implement Callable
func (bo *BinOp) Call(args []Object) (Object, error) {
	if len(args) != 2 {
		return nil, bo.errorf("wrong number of arguments (want 2, got %d)", len(args))
	}

	a, ok := args[0].(float64)
	if !ok {
		return nil, bo.errorf("bad type for first argument - %T", args[0])
	}

	b, ok := args[1].(float64)
	if !ok {
		return nil, bo.errorf("bad type for second argument - %T", args[0])
	}

	val, err := bo.op(a, b)
	if err != nil {
		return nil, bo.errorf("%s", err)
	}

	return val, nil
}

func (bo *BinOp) errorf(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s - %s", bo.name, msg)
}
