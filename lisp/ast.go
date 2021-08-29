package lisp

import (
	"errors"
	"strconv"
)

type Node interface {
	TokenLiteral() string
	IntLiteral() (int64, bool)
	FloatLiteral() (float64, bool)
}

type List struct {
	Nodes []Node
}

func (l List) IntLiteral() (int64, bool) { return 0, false }

func (l List) FloatLiteral() (float64, bool) { return 0, false }

func (l List) TokenLiteral() string {
	str := "( "
	for _, value := range l.Nodes {
		str += value.TokenLiteral() + " "
	}
	str += " )"
	return str
}

type Atom struct {
	Token Token
}
func (a Atom) TokenLiteral() string  {
	return a.Token.Literal
}
func (a Atom) IntLiteral() (int64, bool)  {
	return 0, false
}
func (a Atom) FloatLiteral() (float64, bool)  {
	return 0, false
}

type SymbolAtom struct {
	Atom
	Value string
}

type BoolAtom struct {
	Atom
	Value bool
}

type IntAtom struct {
	Atom
	Value int64
}
func (a IntAtom) IntLiteral() (int64, bool) {
	return a.Value, true
}
func (a IntAtom) FloatLiteral() (float64, bool) {
	return float64(a.Value), true
}

type FloatAtom struct {
	Atom
	Value float64
}
func (a FloatAtom) IntLiteral() (int64, bool) {
	return int64(a.Value), true
}
func (a FloatAtom) FloatLiteral() (float64, bool) {
	return a.Value, true
}

// Callable represents something that is callable....
type Callable func(...Node) (Node, error)
func (a Callable) IntLiteral() (int64, bool) { return 0, false }
func (a Callable) FloatLiteral() (float64, bool) { return 0, false }
func (a Callable) TokenLiteral() string { return "Callable" }


// Scope represents the environment of symbols available to a function
type Scope struct {
	parent *Scope
	objects map[string]Node

}
func NewScope(objects map[string]Node, parent *Scope) *Scope{
	return &Scope{
		parent: parent,
		objects: objects,
	}
}
func (s *Scope) Get(key string) (n Node) {
	n, ok := s.objects[key]
	if !ok && s.parent != nil {
		return s.parent.Get(key)
	}
	return n
}

func (s *Scope) Set(literal string, value Node) {
	s.objects[literal] = value
}

func NewCallable(params List, Body Node, parent *Scope) Callable {
	return func(args ...Node) (Node, error) {
		scope := make(map[string]Node)
		// create args from params list
		for i, param := range params.Nodes {
			scope[param.TokenLiteral()] = args[i]
		}
		env := NewScope(scope, parent)
		return Eval(Body, env)
	}
}


var TokenReadError = errors.New("Token Read Error")

func recursiveReadAST(first Token, rest []Token) (Node, []Token, error){
	if len(rest) == 0 && first.Type == LPAREN || first.Type == RPAREN {
		return nil, nil, TokenReadError
	}
	if first.Type == LPAREN {
		nodes := make([]Node, 0)
		var listTokens = rest
		var n Node
		var err error
		for listTokens[0].Type != RPAREN{
			n, listTokens, err = recursiveReadAST(listTokens[0], listTokens[1:])
			if err != nil {
				return nil, nil, err
			}
			nodes = append(nodes, n)
		}
		return List{Nodes: nodes}, listTokens[1:], nil
	} else if first.Type == RPAREN {
		return nil, nil, TokenReadError
	} else {
		atm, err := atomFromToken(first)
		return atm, rest, err
	}
}

func atomFromToken(t Token) (Node, error) {
	intVal, err := strconv.ParseInt(t.Literal, 10, 64)
	if err != nil {
		floatVal, ferr := strconv.ParseFloat(t.Literal, 64)
		if ferr != nil {
			iv, berr := strconv.ParseBool(t.Literal)
			if berr != nil {
				// return symbol type
				return SymbolAtom{
					Atom:  Atom{ Token: t },
					Value: t.Literal,
				}, nil
			}
			return BoolAtom{
				Atom:  Atom{ Token: t},
				Value: iv,
			}, nil
		}
		return FloatAtom{
			Atom:  Atom{ Token: t },
			Value: floatVal,
		}, nil
	}
	return IntAtom{
		Atom:  Atom{ Token: t },
		Value: intVal,
	}, nil
}

// NewASTFromLex Reads until EOF returning an un-evaluated node
func NewASTFromLex(lexer *Lexer) (Node, error) {
	tokens := make([]Token, 0)
	for {
		tok := lexer.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == EOF {
			break;
		}
	}
	node, _, err := recursiveReadAST(tokens[0], tokens[1:])
	return node, err
}