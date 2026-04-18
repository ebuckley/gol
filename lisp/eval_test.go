package lisp

import (
	"fmt"
	"testing"
)

func TestSimple(t *testing.T) {
	simpleCode := `(+ 1 2)`

	ds := DefaultScope()
	l := NewLexer(simpleCode)
	node, err := NewASTFromLex(l)
	if err != nil {
		t.Fatalf("Should not have errored AST but got %v", err.Error())
	}
	res, err := Eval(node, ds)
	if err != nil {
		t.Fatalf("SHould not have errored eval but got: %v", err.Error())
	}
	t.Log(res.TokenLiteral())
}

func TestIfForm(t *testing.T) {
	simpleCode := `(if true (+ 1 2) (+0 3))`
	ds := DefaultScope()

	l := NewLexer(simpleCode)
	node, err := NewASTFromLex(l)
	if err != nil {
		t.Fatalf("Should not have errored AST but got %v", err.Error())
	}
	res, err := Eval(node, ds)
	if err != nil {
		t.Fatalf("Should not have errored eval but got: %v", err.Error())
	}
	t.Log(res.TokenLiteral())
}

func TestAssignmentForm(t *testing.T) {
	ds := DefaultScope()

	code := `(:= woolygong 89)`
	_, err := EvalString(code, ds)
	if err != nil {
		t.Fatal(err.Error())
	}
	_, foundDef := Environment["woolygong"]
	if !foundDef {
		t.Log("Should have found the defined symbol but did not")
		t.Fail()
	}
}

func TestStringAtom(t *testing.T) {
	ds := DefaultScope()

	res, err := EvalString(`"hello"`, ds)
	if err != nil {
		t.Fatal(err)
	}
	s, ok := res.(StringAtom)
	if !ok {
		t.Fatalf("expected StringAtom, got %T", res)
	}
	if s.Value != "hello" {
		t.Fatalf("expected hello, got %q", s.Value)
	}

	res, err = EvalString(`(= "foo" "foo")`, ds)
	if err != nil {
		t.Fatal(err)
	}
	b, ok := res.(BoolAtom)
	if !ok || !b.Value {
		t.Fatal("expected string equality to return true")
	}

	res, err = EvalString(`(= "foo" "bar")`, ds)
	if err != nil {
		t.Fatal(err)
	}
	b, ok = res.(BoolAtom)
	if !ok || b.Value {
		t.Fatal("expected string inequality to return false")
	}
}

func TestDestructuringAssign(t *testing.T) {
	divide := func(a, b int64) (int64, int64) { return a / b, a % b }
	ds := DefaultScope()
	ds.Set("divmod", GoFunc(divide))

	res, err := EvalString(`(do (:= (q r) (divmod 10 3)) q)`, ds)
	if err != nil {
		t.Fatal(err)
	}
	i, ok := res.IntLiteral()
	if !ok || i != 3 {
		t.Fatalf("expected quotient 3, got %v", res)
	}
}

func TestDestructuringErrorHandling(t *testing.T) {
	failFn := func(x int64) (int64, error) {
		if x < 0 {
			return 0, fmt.Errorf("negative")
		}
		return x * 2, nil
	}
	ds := DefaultScope()
	ds.Set("maybe-double", GoFunc(failFn))

	// success path: val=10, err=nil
	res, err := EvalString(`(do (:= (val err) (maybe-double 5)) val)`, ds)
	if err != nil {
		t.Fatal(err)
	}
	i, ok := res.IntLiteral()
	if !ok || i != 10 {
		t.Fatalf("expected 10, got %v", res)
	}

	// error path: err is a non-nil string, val=0
	res, err = EvalString(`(do (:= (val err) (maybe-double -3)) err)`, ds)
	if err != nil {
		t.Fatal(err)
	}
	errStr, ok := res.(StringAtom)
	if !ok || errStr.Value != "negative" {
		t.Fatalf("expected error string 'negative', got %T %v", res, res)
	}
}

func TestDestructuringNilPadding(t *testing.T) {
	// when the list has fewer values than names, extras get nil
	single := func() int64 { return 42 }
	ds := DefaultScope()
	ds.Set("single", GoFunc(single))

	res, err := EvalString(`(do (:= (a b c) (single)) b)`, ds)
	if err != nil {
		t.Fatal(err)
	}
	b, ok := res.(BoolAtom)
	if !ok || b.Value != false || b.TokenLiteral() != "nil" {
		t.Fatalf("expected nil padding for b, got %T %v", res, res)
	}
}

func TestAssignmentUsage(t *testing.T) {
	ds := DefaultScope()

	code := `(do (:= woolygong 89) (+ woolygong 1))`
	res, err := EvalString(code, ds)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Log("map of state yo", Environment)
	_, foundDef := Environment["woolygong"]
	if !foundDef {
		t.Log("Should have found the defined symbol but did not")
		t.Fail()
	}
	lit := res.TokenLiteral()
	if lit != "90" {
		t.Log("result is:", res.TokenLiteral())
		t.Fail()
	}
}