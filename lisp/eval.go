package lisp

import (
	"errors"
	"fmt"
)



var Environment = map[string]Node {
	"*": Callable(func(node ...Node) (Node, error) {
		lhs, lOk := node[0].FloatLiteral()
		rhs, rOk := node[1].FloatLiteral()
		if !lOk {
			return nil, errors.New("Could not find float value for: " + node[0].TokenLiteral())
		}
		if !rOk {
			return nil, errors.New("Could not find float value for: " + node[1].TokenLiteral())
		}
		res := lhs * rhs
		return FloatAtom{
			Atom:  Atom{ Token: Token{
				Type: SYMBOL, Literal: fmt.Sprint(res)}},
			Value: lhs * rhs,
		}, nil
	}),
	"+": Callable(func(node ...Node) (Node, error) {


		var currVal float64 = 0
		for _, n := range node {
			currFloat, anyIsFloat := n.(FloatAtom)
			if anyIsFloat {
				currVal += currFloat.Value
			}
			num, isInt := n.(IntAtom)
			if isInt {
				currVal += float64(num.Value)
			}
		}
		resultAtom := FloatAtom{
			Atom:  Atom{ Token: Token{
				Type:    SYMBOL,
				Literal: fmt.Sprint(currVal),
			} },
			Value: currVal,
		}
		return resultAtom, nil
	}),
}

func DefaultScope() *Scope {
	return &Scope{
		parent:  nil,
		objects: Environment,
	}
}

func Eval(n Node, scope *Scope) (Node, error) {

	symbol, isSymbol := n.(SymbolAtom)
	if isSymbol {
		cb := scope.Get(symbol.TokenLiteral())
		if cb == nil {
			return nil, errors.New("symbol not found: " +  symbol.TokenLiteral())
		}
		return cb, nil
	}
	float, isFloat := n.(FloatAtom)
	if isFloat {
		return float, nil
	}
	num, isInt := n.(IntAtom)
	if isInt {
		return num, nil
	}

	if b, isBool := n.(BoolAtom); isBool {
		return b, nil
	}

	// Function application happens.
	L, isList := n.(List)
	if isList {
		fst := L.Nodes[0]

		// IF is a special case form
		if fst.TokenLiteral() == "if" {
			cond, err := Eval(L.Nodes[1], scope)
			if err != nil {
				return nil, err
			}
			b, ok := cond.(BoolAtom)
			if !ok {
				return nil, errors.New("if form expected a boolean result but got " + cond.TokenLiteral())
			}
			if b.Value {
				return Eval(L.Nodes[2], scope)
			} else {
				if len(L.Nodes) == 4 {
					return Eval(L.Nodes[3],scope)
				} else {
					return nil, nil
				}
			}
		} else if fst.TokenLiteral() == ":=" {

			name, isSymbol := L.Nodes[1].(SymbolAtom)
			if !isSymbol {
				return nil, errors.New("error with special form := " +  L.TokenLiteral())
			}
			value, err := Eval(L.Nodes[2], scope)
			if err != nil {
				return nil, err
			}
			// map the first param name to be the value in the current environment
			scope.Set(name.TokenLiteral(), value)
			return name, nil
		} else if fst.TokenLiteral() == "do" {
			// (do (x) (y) (z)) is a special form which iteratively Evals the enclosed forms, resulting in the final form being returned
			var result Node
			var rErr error
			for _, n := range L.Nodes[1:] {
				result, rErr =  Eval(n, scope)
				if rErr != nil {
					return result, rErr
				}
			}
			return result, nil
		} else if fst.TokenLiteral() == "func" {
			// a func is a special form (func (x) (+ x 1))
			params, ok := L.Nodes[1].(List)
			if !ok {
				return nil, errors.New("special form func requires a list of params as the second parameter")
			}
			body := L.Nodes[2]
			return NewCallable(params, body, scope), nil
		} else if fst.TokenLiteral() == "quote" {
			return L.Nodes[1], nil
		} else if fst.TokenLiteral() == "set!" {
			symbol := L.Nodes[1]
			expr := L.Nodes[2]
			r, err := Eval(expr, scope)
			if err != nil {
				return nil, err
			}
			scope.Set(symbol.TokenLiteral(), r)
			return r, nil
		}

		maybeCallable, err := Eval(fst, scope) // should return a callable of some type from the environment...
		if err != nil {
			return nil, err
		}
		proc, ok := maybeCallable.(Callable)
		if !ok {
			return nil, errors.New(fmt.Sprintf("the first element in a list should be a callable node but we got: %s", fst.TokenLiteral()))
		}
		args := make([]Node, 0)
		for _, node := range L.Nodes[1:] {
			iNode, err := Eval(node, scope)
			if err != nil {
				return iNode, err
			}
			args = append(args, iNode)
		}
		// Evaluate and return the nodes
		return proc(args...)
	}
	return nil, errors.New("fatal internal evaluation error") // not possible?
}

func EvalString(code string, scope *Scope) (Node, error) {
	l := NewLexer(code)
	node, err := NewASTFromLex(l)
	if err != nil {
		return node, fmt.Errorf("AST parser error: %v", err.Error())
	}
	return Eval(node, scope)
}