package lisp

import "testing"

func TestHashMapBasic(t *testing.T) {
	scope := DefaultScope()

	// empty hash map
	n, err := EvalString(`(hash-map)`, scope)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := n.(HashMapNode); !ok {
		t.Fatalf("expected HashMapNode, got %T", n)
	}

	// hash-map with initial pairs
	n, err = EvalString(`(hash-map "a" 1 "b" 2)`, scope)
	if err != nil {
		t.Fatal(err)
	}
	h, ok := n.(HashMapNode)
	if !ok {
		t.Fatalf("expected HashMapNode, got %T", n)
	}
	if len(h.m.Entries()) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(h.m.Entries()))
	}
}

func TestHashMapAssocGet(t *testing.T) {
	scope := DefaultScope()

	n, err := EvalString(`(hm-get (hm-assoc (hash-map) "x" 42) "x")`, scope)
	if err != nil {
		t.Fatal(err)
	}
	i, ok := n.(IntAtom)
	if !ok {
		t.Fatalf("expected IntAtom, got %T", n)
	}
	if i.Value != 42 {
		t.Fatalf("expected 42, got %d", i.Value)
	}
}

func TestHashMapImmutability(t *testing.T) {
	scope := DefaultScope()

	// original map unchanged after assoc
	_, err := EvalProgram(`
(:= m (hash-map "a" 1))
(:= m2 (hm-assoc m "b" 2))
`, scope)
	if err != nil {
		t.Fatal(err)
	}

	orig := scope.Get("m").(HashMapNode)
	if len(orig.m.Entries()) != 1 {
		t.Fatalf("original map should still have 1 entry, got %d", len(orig.m.Entries()))
	}

	updated := scope.Get("m2").(HashMapNode)
	if len(updated.m.Entries()) != 2 {
		t.Fatalf("updated map should have 2 entries, got %d", len(updated.m.Entries()))
	}
}

func TestHashMapDissoc(t *testing.T) {
	scope := DefaultScope()

	n, err := EvalString(`(hm-count (hm-dissoc (hash-map "a" 1 "b" 2) "a"))`, scope)
	if err != nil {
		t.Fatal(err)
	}
	i, ok := n.(IntAtom)
	if !ok {
		t.Fatalf("expected IntAtom, got %T", n)
	}
	if i.Value != 1 {
		t.Fatalf("expected 1, got %d", i.Value)
	}
}

func TestHashMapGetMissing(t *testing.T) {
	scope := DefaultScope()

	n, err := EvalString(`(hm-get (hash-map) "missing")`, scope)
	if err != nil {
		t.Fatal(err)
	}
	b, ok := n.(BoolAtom)
	if !ok || b.TokenLiteral() != "nil" {
		t.Fatalf("expected nil for missing key, got %v", n)
	}
}

func TestHashMapEntries(t *testing.T) {
	scope := DefaultScope()

	n, err := EvalString(`(hm-entries (hash-map "k" 99))`, scope)
	if err != nil {
		t.Fatal(err)
	}
	list, ok := n.(List)
	if !ok {
		t.Fatalf("expected List, got %T", n)
	}
	if len(list.Nodes) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(list.Nodes))
	}
	pair, ok := list.Nodes[0].(List)
	if !ok || len(pair.Nodes) != 2 {
		t.Fatalf("expected pair list")
	}
	if pair.Nodes[0].(StringAtom).Value != "k" {
		t.Fatalf("expected key 'k'")
	}
}
