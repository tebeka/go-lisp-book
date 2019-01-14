package main

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
