package lisp

import "fmt"

// RegisterFunc registers a set of symbols into a scope under a given namespace.
type RegisterFunc func(scope *Scope)

var globalRegistry = map[string]RegisterFunc{}

// Register adds a package registrar to the global registry under name.
// Call this from init() in generated binding packages.
func Register(name string, fn RegisterFunc) {
	globalRegistry[name] = fn
}

// ImportInto loads a registered package into scope as a namespace.
// It is the runtime implementation of the (import name) special form.
func ImportInto(name string, scope *Scope) error {
	fn, ok := globalRegistry[name]
	if !ok {
		return fmt.Errorf("import: package %q not found in registry", name)
	}
	ns := NewScope(map[string]Node{}, nil)
	fn(ns)
	scope.Set(name, ObjectNode{Fields: ns.objects})
	return nil
}
