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

var (
	nilNode = BoolAtom{Atom: Atom{Token: Token{Type: SYMBOL, Literal: "nil"}}}
	errType = reflect.TypeOf((*error)(nil)).Elem()
	anyType = reflect.TypeOf((*any)(nil)).Elem()
	nodeType = reflect.TypeOf((*Node)(nil)).Elem()
)

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

// buildCallArgs validates arg count and coerces Node args to reflect.Values.
func buildCallArgs(name string, args []Node, ft reflect.Type) ([]reflect.Value, error) {
	numFixed := ft.NumIn()
	if ft.IsVariadic() {
		numFixed--
		if len(args) < numFixed {
			return nil, fmt.Errorf("%s: expected at least %d args, got %d", name, numFixed, len(args))
		}
	} else if numFixed != len(args) {
		return nil, fmt.Errorf("%s: expected %d args, got %d", name, numFixed, len(args))
	}
	in := make([]reflect.Value, len(args))
	for i := 0; i < numFixed; i++ {
		v, err := nodeToReflect(args[i], ft.In(i))
		if err != nil {
			return nil, fmt.Errorf("%s arg %d: %w", name, i, err)
		}
		in[i] = v
	}
	if ft.IsVariadic() {
		elemType := ft.In(ft.NumIn() - 1).Elem()
		for i := numFixed; i < len(args); i++ {
			v, err := nodeToReflect(args[i], elemType)
			if err != nil {
				return nil, fmt.Errorf("%s arg %d: %w", name, i, err)
			}
			in[i] = v
		}
	}
	return in, nil
}

// reflectCallable wraps a reflect.Value (must be a Func) as a Callable.
func reflectCallable(name string, fn reflect.Value) Callable {
	ft := fn.Type()
	return func(args ...Node) (Node, error) {
		in, err := buildCallArgs(name, args, ft)
		if err != nil {
			return nil, err
		}
		return reflectResultToNode(name, fn.Call(in), ft)
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
	case reflect.Slice:
		if list, ok := n.(List); ok {
			elemType := target.Elem()
			slice := reflect.MakeSlice(target, len(list.Nodes), len(list.Nodes))
			for i, item := range list.Nodes {
				v, err := nodeToReflect(item, elemType)
				if err != nil {
					return reflect.Value{}, fmt.Errorf("slice elem %d: %w", i, err)
				}
				slice.Index(i).Set(v)
			}
			return slice, nil
		}
	case reflect.Interface:
		if g, ok := n.(GoValue); ok {
			rv := reflect.ValueOf(g.Value)
			if rv.Type().Implements(target) {
				return rv, nil
			}
		}
		// pass any node as interface{}: unwrap to native Go value
		if target == anyType {
			switch v := n.(type) {
			case StringAtom:
				return reflect.ValueOf(v.Value), nil
			case IntAtom:
				return reflect.ValueOf(v.Value), nil
			case FloatAtom:
				return reflect.ValueOf(v.Value), nil
			case BoolAtom:
				return reflect.ValueOf(v.Value), nil
			case GoValue:
				return reflect.ValueOf(v.Value), nil
			default:
				return reflect.ValueOf(n), nil
			}
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
// When the function signature ends with error, the error is always included as
// the last element of the returned List (as nil or a StringAtom), so lisp code
// can destructure (:= (val err) (some-fn ...)) without exceptions.
func reflectResultToNode(name string, out []reflect.Value, ft reflect.Type) (Node, error) {
	hasErrReturn := ft.NumOut() > 0 && ft.Out(ft.NumOut()-1).Implements(errType)

	if !hasErrReturn {
		if len(out) == 0 {
			return nilNode, nil
		}
		if len(out) == 1 {
			return goValueToNode(out[0]), nil
		}
		nodes := make([]Node, len(out))
		for i, v := range out {
			nodes[i] = goValueToNode(v)
		}
		return List{Nodes: nodes}, nil
	}

	// functions with error return: always surface as (value... err)
	errVal := out[len(out)-1]
	valueOuts := out[:len(out)-1]

	var errNode Node
	if errVal.IsNil() {
		errNode = nilNode
	} else {
		msg := errVal.Interface().(error).Error()
		errNode = StringAtom{Atom: Atom{Token: Token{Type: STRING, Literal: msg}}, Value: msg}
	}

	if len(valueOuts) == 0 {
		return List{Nodes: []Node{nilNode, errNode}}, nil
	}
	nodes := make([]Node, len(valueOuts)+1)
	for i, v := range valueOuts {
		nodes[i] = goValueToNode(v)
	}
	nodes[len(valueOuts)] = errNode
	return List{Nodes: nodes}, nil
}

// goValueToNode converts a reflect.Value to the most specific Node type possible.
func goValueToNode(rv reflect.Value) Node {
	if !rv.IsValid() {
		return nilNode
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
	case reflect.Slice:
		if rv.IsNil() {
			return nilNode
		}
		nodes := make([]Node, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			nodes[i] = goValueToNode(rv.Index(i))
		}
		return List{Nodes: nodes}
	case reflect.Pointer:
		if rv.IsNil() {
			return nilNode
		}
		obj, err := WrapObject(rv.Interface())
		if err != nil {
			return GoValue{Value: rv.Interface()}
		}
		return obj
	}
	return GoValue{Value: rv.Interface()}
}
