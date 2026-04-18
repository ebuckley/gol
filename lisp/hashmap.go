package lisp

import (
	"fmt"
	"strings"

	"hmt"
)

// HashMapNode wraps an immutable hmt.HMT as a lisp Node.
type HashMapNode struct {
	m *hmt.HMT[Node]
}

func newHashMapNode() HashMapNode {
	return HashMapNode{m: hmt.New[Node]()}
}

func (h HashMapNode) TokenLiteral() string {
	entries := h.m.Entries()
	parts := make([]string, 0, len(entries))
	for _, e := range entries {
		parts = append(parts, fmt.Sprintf("%s %s", string(e.Key), e.Value.TokenLiteral()))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
func (h HashMapNode) IntLiteral() (int64, bool)    { return 0, false }
func (h HashMapNode) FloatLiteral() (float64, bool) { return 0, false }

func hmAssoc(h HashMapNode, key string, val Node) (HashMapNode, error) {
	m2, err := h.m.Set(hmt.Key(key), val)
	if err != nil {
		return h, err
	}
	return HashMapNode{m: m2}, nil
}

func hmDissoc(h HashMapNode, key string) (HashMapNode, error) {
	m2, err := h.m.Del(hmt.Key(key))
	if err != nil {
		return h, err
	}
	return HashMapNode{m: m2}, nil
}

func hmGet(h HashMapNode, key string) (Node, error) {
	entry, err := h.m.Get(hmt.Key(key))
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nilNode, nil
	}
	return entry.Value, nil
}

func init() {
	for k, v := range hashMapBuiltins {
		Environment[k] = v
	}
}

// hashMapBuiltins are the builtin functions for hash maps.
var hashMapBuiltins = map[string]Node{
	// (hash-map) or (hash-map "k" v "k2" v2 ...)
	"hash-map": Callable(func(args ...Node) (Node, error) {
		if len(args)%2 != 0 {
			return nil, fmt.Errorf("hash-map: expected even number of args, got %d", len(args))
		}
		h := newHashMapNode()
		for i := 0; i < len(args); i += 2 {
			key, ok := args[i].(StringAtom)
			if !ok {
				return nil, fmt.Errorf("hash-map: keys must be strings, got %T", args[i])
			}
			var err error
			h, err = hmAssoc(h, key.Value, args[i+1])
			if err != nil {
				return nil, err
			}
		}
		return h, nil
	}),

	// (hm-assoc m "key" val) → new map
	"hm-assoc": Callable(func(args ...Node) (Node, error) {
		if len(args) != 3 {
			return nil, fmt.Errorf("hm-assoc: expected (hm-assoc map key val)")
		}
		h, ok := args[0].(HashMapNode)
		if !ok {
			return nil, fmt.Errorf("hm-assoc: first arg must be a hash-map")
		}
		key, ok := args[1].(StringAtom)
		if !ok {
			return nil, fmt.Errorf("hm-assoc: key must be a string")
		}
		return hmAssoc(h, key.Value, args[2])
	}),

	// (hm-dissoc m "key") → new map without key
	"hm-dissoc": Callable(func(args ...Node) (Node, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("hm-dissoc: expected (hm-dissoc map key)")
		}
		h, ok := args[0].(HashMapNode)
		if !ok {
			return nil, fmt.Errorf("hm-dissoc: first arg must be a hash-map")
		}
		key, ok := args[1].(StringAtom)
		if !ok {
			return nil, fmt.Errorf("hm-dissoc: key must be a string")
		}
		return hmDissoc(h, key.Value)
	}),

	// (hm-get m "key") → value or nil
	"hm-get": Callable(func(args ...Node) (Node, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("hm-get: expected (hm-get map key)")
		}
		h, ok := args[0].(HashMapNode)
		if !ok {
			return nil, fmt.Errorf("hm-get: first arg must be a hash-map")
		}
		key, ok := args[1].(StringAtom)
		if !ok {
			return nil, fmt.Errorf("hm-get: key must be a string")
		}
		return hmGet(h, key.Value)
	}),

	// (hm-entries m) → list of ("key" val) pairs
	"hm-entries": Callable(func(args ...Node) (Node, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("hm-entries: expected (hm-entries map)")
		}
		h, ok := args[0].(HashMapNode)
		if !ok {
			return nil, fmt.Errorf("hm-entries: arg must be a hash-map")
		}
		entries := h.m.Entries()
		pairs := make([]Node, len(entries))
		for i, e := range entries {
			k := string(e.Key)
			pairs[i] = List{Nodes: []Node{
				StringAtom{Atom: Atom{Token: Token{Type: STRING, Literal: k}}, Value: k},
				e.Value,
			}}
		}
		return List{Nodes: pairs}, nil
	}),

	// (hm-count m) → int
	"hm-count": Callable(func(args ...Node) (Node, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("hm-count: expected (hm-count map)")
		}
		h, ok := args[0].(HashMapNode)
		if !ok {
			return nil, fmt.Errorf("hm-count: arg must be a hash-map")
		}
		n := int64(len(h.m.Entries()))
		return IntAtom{Atom: Atom{Token: Token{Type: SYMBOL, Literal: fmt.Sprint(n)}}, Value: n}, nil
	}),
}
