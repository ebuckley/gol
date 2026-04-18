package lisp

import (
	"fmt"
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

	res, err := EvalString(`(maybe-double 5)`, ds)
	if err != nil {
		t.Fatal(err)
	}
	i, ok := res.(IntAtom)
	if !ok || i.Value != 10 {
		t.Fatalf("expected IntAtom 10, got %T %v", res, res)
	}

	_, err = EvalString(`(maybe-double -1)`, ds)
	if err == nil {
		t.Fatal("expected error for negative input")
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
