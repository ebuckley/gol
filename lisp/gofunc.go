package lisp

import (
	"errors"
	"fmt"
	"reflect"
)

// GoFunc wraps any Go function as a Callable, auto-coercing Node args to Go
// types and converting return values back to Node.
func GoFunc(fn any) Callable {
	fv := reflect.ValueOf(fn)
	if fv.Kind() != reflect.Func {
		panic("GoFunc: argument must be a function")
	}
	ft := fv.Type()
	name := fmt.Sprintf("%T", fn)
	return func(args ...Node) (Node, error) {
		in, err := buildCallArgs(name, args, ft)
		if err != nil {
			return nil, err
		}
		return reflectResultToNode(name, fv.Call(in), ft)
	}
}

// ToNode converts a plain Go value to its most specific Node representation.
// Useful when returning values from hand-written Callables.
func ToNode(v any) Node {
	if v == nil {
		return nilNode
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
