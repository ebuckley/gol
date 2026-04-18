package lisp

import (
	"errors"
	"fmt"
	"reflect"
)

// ObjectNode holds eagerly-extracted fields and methods from a Go value.
type ObjectNode struct {
	Fields map[string]Node
}

func (o ObjectNode) TokenLiteral() string     { return fmt.Sprintf("Object%v", o.Fields) }
func (o ObjectNode) IntLiteral() (int64, bool)    { return 0, false }
func (o ObjectNode) FloatLiteral() (float64, bool) { return 0, false }

// WrapObject uses reflection to build an ObjectNode from any Go struct or pointer.
// Exported fields become GoValue nodes; exported methods become Callable nodes.
func WrapObject(v any) (ObjectNode, error) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return ObjectNode{}, errors.New("WrapObject: nil value")
	}

	fields := make(map[string]Node)

	// methods on the value (including pointer receiver methods)
	rt := rv.Type()
	for i := 0; i < rt.NumMethod(); i++ {
		m := rt.Method(i)
		if !m.IsExported() {
			continue
		}
		mv := rv.Method(i)
		fields[m.Name] = reflectCallable(m.Name, mv)
	}

	// fields on the underlying struct
	re := rv
	if re.Kind() == reflect.Pointer {
		re = re.Elem()
	}
	if re.Kind() == reflect.Struct {
		ret := re.Type()
		for i := 0; i < ret.NumField(); i++ {
			f := ret.Field(i)
			if !f.IsExported() {
				continue
			}
			fields[f.Name] = goValueToNode(re.Field(i))
		}
	}

	return ObjectNode{Fields: fields}, nil
}

// reflectCallable wraps a reflect.Value (must be a Func) as a Callable.
func reflectCallable(name string, fn reflect.Value) Callable {
	return func(args ...Node) (Node, error) {
		ft := fn.Type()
		if ft.IsVariadic() {
			return nil, fmt.Errorf("%s: variadic methods not supported", name)
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

		out := fn.Call(in)
		return reflectResultToNode(name, out, fn.Type())
	}
}

// nodeToReflect coerces a Node to the target reflect.Type.
func nodeToReflect(n Node, target reflect.Type) (reflect.Value, error) {
	switch target.Kind() {
	case reflect.String:
		if s, ok := n.(StringAtom); ok {
			return reflect.ValueOf(s.Value), nil
		}
		return reflect.ValueOf(n.TokenLiteral()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, ok := n.IntLiteral(); ok {
			return reflect.ValueOf(i).Convert(target), nil
		}
	case reflect.Float32, reflect.Float64:
		if f, ok := n.FloatLiteral(); ok {
			return reflect.ValueOf(f).Convert(target), nil
		}
	case reflect.Bool:
		if b, ok := n.(BoolAtom); ok {
			return reflect.ValueOf(b.Value), nil
		}
	case reflect.Interface:
		if g, ok := n.(GoValue); ok {
			rv := reflect.ValueOf(g.Value)
			if rv.Type().Implements(target) {
				return rv, nil
			}
		}
		// pass any node as interface{}
		if target == reflect.TypeOf((*any)(nil)).Elem() {
			if g, ok := n.(GoValue); ok {
				return reflect.ValueOf(g.Value), nil
			}
			return reflect.ValueOf(n), nil
		}
	default:
		if g, ok := n.(GoValue); ok {
			rv := reflect.ValueOf(g.Value)
			if rv.Type().AssignableTo(target) {
				return rv, nil
			}
			if rv.Type().ConvertibleTo(target) {
				return rv.Convert(target), nil
			}
		}
	}
	return reflect.Value{}, fmt.Errorf("cannot coerce %T to %v", n, target)
}

// reflectResultToNode converts reflect call output to a Node.
func reflectResultToNode(name string, out []reflect.Value, ft reflect.Type) (Node, error) {
	errType := reflect.TypeOf((*error)(nil)).Elem()

	// check last return for error
	if len(out) > 0 && ft.Out(ft.NumOut()-1).Implements(errType) {
		last := out[len(out)-1]
		if !last.IsNil() {
			return nil, last.Interface().(error)
		}
		out = out[:len(out)-1]
	}

	if len(out) == 0 {
		return BoolAtom{Atom: Atom{Token: Token{Type: SYMBOL, Literal: "nil"}}}, nil
	}
	return goValueToNode(out[0]), nil
}

var nodeType = reflect.TypeOf((*Node)(nil)).Elem()

// goValueToNode converts a reflect.Value to the most specific Node type possible.
func goValueToNode(rv reflect.Value) Node {
	if !rv.IsValid() {
		return BoolAtom{Atom: Atom{Token: Token{Type: SYMBOL, Literal: "nil"}}}
	}
	if rv.Type().Implements(nodeType) {
		return rv.Interface().(Node)
	}
	switch rv.Kind() {
	case reflect.String:
		s := rv.String()
		return StringAtom{Atom: Atom{Token: Token{Type: STRING, Literal: s}}, Value: s}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i := rv.Int()
		return IntAtom{Atom: Atom{Token: Token{Type: SYMBOL, Literal: fmt.Sprint(i)}}, Value: i}
	case reflect.Float32, reflect.Float64:
		f := rv.Float()
		return FloatAtom{Atom: Atom{Token: Token{Type: SYMBOL, Literal: fmt.Sprint(f)}}, Value: f}
	case reflect.Bool:
		b := rv.Bool()
		return BoolAtom{Atom: Atom{Token: Token{Type: SYMBOL, Literal: fmt.Sprint(b)}}, Value: b}
	case reflect.Struct:
		obj, err := WrapObject(rv.Interface())
		if err != nil {
			return GoValue{Value: rv.Interface()}
		}
		return obj
	case reflect.Pointer:
		if rv.IsNil() {
			return BoolAtom{Atom: Atom{Token: Token{Type: SYMBOL, Literal: "nil"}}}
		}
		obj, err := WrapObject(rv.Interface())
		if err != nil {
			return GoValue{Value: rv.Interface()}
		}
		return obj
	}
	return GoValue{Value: rv.Interface()}
}
