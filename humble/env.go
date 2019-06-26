package main

// Environment holds name → values
type Environment struct {
	bindings map[string]Object
	parent   *Environment
}

// NewEnvironment returns a new environment
func NewEnvironment(bindings map[string]Object, parent *Environment) *Environment {
	return &Environment{bindings, parent}
}

// Find finds the environment holding name, return nil if not found
func (e *Environment) Find(name string) *Environment {
	if _, ok := e.bindings[name]; ok {
		return e
	}

	if e.parent == nil {
		return nil
	}
	return e.parent.Find(name)
}

// Get returns bindings for name in environment
func (e *Environment) Get(name string) Object {
	return e.bindings[name]
}

// Set sets bindings for name
func (e *Environment) Set(name string, value Object) {
	e.bindings[name] = value
}
