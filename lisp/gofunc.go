package lisp

import (
	"errors"
	"fmt"
	"reflect"
)

// GoFunc wraps any Go function as a Callable, auto-coercing Node args to Go
// types and converting return values back to Node.
func GoFunc(fn any) Callable {
	rv := reflect.ValueOf(fn)
	if rv.Kind() != reflect.Func {
		panic("GoFunc: argument must be a function")
	}
	ft := rv.Type()
	name := fmt.Sprintf("%T", fn)
	return func(args ...Node) (Node, error) {
		if ft.IsVariadic() {
			return nil, fmt.Errorf("%s: variadic functions not supported via GoFunc", name)
		}
		if ft.NumIn() != len(args) {
			return nil, fmt.Errorf("%s: expected %d args, got %d", name, ft.NumIn(), len(args))
		}
		in := make([]reflect.Value, len(args))
		for i, arg := range args {
			want := ft.In(i)
			rv, err := nodeToReflect(arg, want)
			if err != nil {
				return nil, fmt.Errorf("%s arg %d: %w", name, i, err)
			}
			in[i] = rv
		}
		out := rv.Call(in)
		return reflectResultToNode(name, out, ft)
	}
}

// ToNode converts a plain Go value to its most specific Node representation.
// Useful when returning values from hand-written Callables.
func ToNode(v any) Node {
	if v == nil {
		return BoolAtom{Atom: Atom{Token: Token{Type: SYMBOL, Literal: "nil"}}}
	}
	return goValueToNode(reflect.ValueOf(v))
}

// FromNode extracts a typed Go value from a Node, returning an error if the
// conversion is not possible.
func FromNode[T any](n Node) (T, error) {
	var zero T
	target := reflect.TypeOf(&zero).Elem()
	rv, err := nodeToReflect(n, target)
	if err != nil {
		return zero, err
	}
	v, ok := rv.Interface().(T)
	if !ok {
		return zero, errors.New("FromNode: type assertion failed")
	}
	return v, nil
}
