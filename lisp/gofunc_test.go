package lisp

import (
	"fmt"
	"strings"
	"testing"
)

func TestGoFuncBasic(t *testing.T) {
	add := func(a, b int64) int64 { return a + b }

	ds := DefaultScope()
	ds.Set("go-add", GoFunc(add))

	res, err := EvalString(`(go-add 3 4)`, ds)
	if err != nil {
		t.Fatal(err)
	}
	i, ok := res.(IntAtom)
	if !ok || i.Value != 7 {
		t.Fatalf("expected IntAtom 7, got %T %v", res, res)
	}
}

func TestGoFuncStringReturn(t *testing.T) {
	greet := func(name string) string { return "Hello " + name }

	ds := DefaultScope()
	ds.Set("greet", GoFunc(greet))

	res, err := EvalString(`(greet "World")`, ds)
	if err != nil {
		t.Fatal(err)
	}
	s, ok := res.(StringAtom)
	if !ok || s.Value != "Hello World" {
		t.Fatalf("expected StringAtom 'Hello World', got %T %v", res, res)
	}
}

func TestGoFuncErrorReturn(t *testing.T) {
	failFn := func(x int64) (int64, error) {
		if x < 0 {
			return 0, fmt.Errorf("negative: %d", x)
		}
		return x * 2, nil
	}

	ds := DefaultScope()
	ds.Set("maybe-double", GoFunc(failFn))

	// success: returns (val nil) list
	res, err := EvalString(`(maybe-double 5)`, ds)
	if err != nil {
		t.Fatal(err)
	}
	list, ok := res.(List)
	if !ok || len(list.Nodes) != 2 {
		t.Fatalf("expected List of 2, got %T %v", res, res)
	}
	i, ok := list.Nodes[0].IntLiteral()
	if !ok || i != 10 {
		t.Fatalf("expected val=10, got %v", list.Nodes[0])
	}
	if _, isNil := list.Nodes[1].(BoolAtom); !isNil {
		t.Fatalf("expected nil err, got %T %v", list.Nodes[1], list.Nodes[1])
	}

	// failure: returns (0 "negative: -1") list — no Go-level panic
	res, err = EvalString(`(maybe-double -1)`, ds)
	if err != nil {
		t.Fatal(err)
	}
	list, ok = res.(List)
	if !ok || len(list.Nodes) != 2 {
		t.Fatalf("expected List of 2 on error, got %T %v", res, res)
	}
	errNode, ok := list.Nodes[1].(StringAtom)
	if !ok || errNode.Value != "negative: -1" {
		t.Fatalf("expected error string, got %T %v", list.Nodes[1], list.Nodes[1])
	}
}

func TestGoFuncVariadic(t *testing.T) {
	ds := DefaultScope()
	ds.Set("sprintf", GoFunc(fmt.Sprintf))
	ds.Set("join", GoFunc(strings.Join))

	res, err := EvalString(`(sprintf "%s=%d" "x" 42)`, ds)
	if err != nil {
		t.Fatal(err)
	}
	s, ok := res.(StringAtom)
	if !ok || s.Value != "x=42" {
		t.Fatalf("expected 'x=42', got %T %v", res, res)
	}

	// strings.Join takes ([]string, sep) — first arg is a slice
	res2, err := EvalString(`(join (quote ("a" "b" "c")) "-")`, ds)
	if err != nil {
		t.Fatal(err)
	}
	s2, ok := res2.(StringAtom)
	if !ok || s2.Value != "a-b-c" {
		t.Fatalf("expected 'a-b-c', got %T %v", res2, res2)
	}
}

func TestGoFuncMultiReturn(t *testing.T) {
	divide := func(a, b int64) (int64, int64) { return a / b, a % b }
	ds := DefaultScope()
	ds.Set("divmod", GoFunc(divide))

	res, err := EvalString(`(divmod 10 3)`, ds)
	if err != nil {
		t.Fatal(err)
	}
	list, ok := res.(List)
	if !ok || len(list.Nodes) != 2 {
		t.Fatalf("expected List of 2, got %T %v", res, res)
	}
	q, _ := list.Nodes[0].IntLiteral()
	r, _ := list.Nodes[1].IntLiteral()
	if q != 3 || r != 1 {
		t.Fatalf("expected (3 1), got (%d %d)", q, r)
	}
}

func TestSliceRoundtrip(t *testing.T) {
	ds := DefaultScope()
	ds.Set("split", GoFunc(strings.Split))

	res, err := EvalString(`(split "a,b,c" ",")`, ds)
	if err != nil {
		t.Fatal(err)
	}
	list, ok := res.(List)
	if !ok || len(list.Nodes) != 3 {
		t.Fatalf("expected List of 3, got %T %v", res, res)
	}
}

func TestNamespaceScope(t *testing.T) {
	ds := DefaultScope()
	ds.SetNamespace("str", map[string]Node{
		"contains": GoFunc(strings.Contains),
		"upper":    GoFunc(strings.ToUpper),
	})

	res, err := EvalString(`(str/contains "hello world" "world")`, ds)
	if err != nil {
		t.Fatal(err)
	}
	b, ok := res.(BoolAtom)
	if !ok || !b.Value {
		t.Fatalf("expected true, got %T %v", res, res)
	}

	res2, err := EvalString(`(str/upper "hello")`, ds)
	if err != nil {
		t.Fatal(err)
	}
	s, ok := res2.(StringAtom)
	if !ok || s.Value != "HELLO" {
		t.Fatalf("expected 'HELLO', got %T %v", res2, res2)
	}
}

func TestToNodeFromNode(t *testing.T) {
	n := ToNode("hello")
	s, ok := n.(StringAtom)
	if !ok || s.Value != "hello" {
		t.Fatalf("expected StringAtom hello, got %T", n)
	}

	got, err := FromNode[string](n)
	if err != nil || got != "hello" {
		t.Fatalf("expected hello, got %q err %v", got, err)
	}

	n2 := ToNode(int64(42))
	i, ok := n2.(IntAtom)
	if !ok || i.Value != 42 {
		t.Fatalf("expected IntAtom 42, got %T", n2)
	}
}
