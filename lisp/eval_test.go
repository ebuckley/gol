package lisp

import "testing"

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