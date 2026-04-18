package lisp

import "testing"

type testPerson struct {
	Name string
	Age  int
}

func (p testPerson) Greet(suffix string) string {
	return "Hello " + p.Name + suffix
}

func (p testPerson) Double() int {
	return p.Age * 2
}

func TestWrapObject(t *testing.T) {
	p := testPerson{Name: "Alice", Age: 30}
	obj, err := WrapObject(p)
	if err != nil {
		t.Fatal(err)
	}

	nameNode, ok := obj.Fields["Name"]
	if !ok {
		t.Fatal("expected Name field")
	}
	s, ok := nameNode.(StringAtom)
	if !ok || s.Value != "Alice" {
		t.Fatalf("expected StringAtom Alice, got %T %v", nameNode, nameNode)
	}

	ageNode, ok := obj.Fields["Age"]
	if !ok {
		t.Fatal("expected Age field")
	}
	i, ok := ageNode.(IntAtom)
	if !ok || i.Value != 30 {
		t.Fatalf("expected IntAtom 30, got %T %v", ageNode, ageNode)
	}

	greetFn, ok := obj.Fields["Greet"].(Callable)
	if !ok {
		t.Fatal("expected Greet to be Callable")
	}
	result, err := greetFn(StringAtom{Atom: Atom{Token: Token{Type: STRING, Literal: "!"}}, Value: "!"})
	if err != nil {
		t.Fatal(err)
	}
	if result.TokenLiteral() != "Hello Alice!" {
		t.Fatalf("expected 'Hello Alice!', got %q", result.TokenLiteral())
	}
}

func TestGetBuiltin(t *testing.T) {
	ds := DefaultScope()
	p := testPerson{Name: "Bob", Age: 25}
	obj, err := WrapObject(p)
	if err != nil {
		t.Fatal(err)
	}
	ds.Set("p", obj)

	res, err := EvalString(`(get p "Name")`, ds)
	if err != nil {
		t.Fatal(err)
	}
	s, ok := res.(StringAtom)
	if !ok || s.Value != "Bob" {
		t.Fatalf("expected StringAtom Bob, got %T %v", res, res)
	}
}

func TestGetAndCallMethod(t *testing.T) {
	ds := DefaultScope()
	p := testPerson{Name: "Carol", Age: 20}
	obj, err := WrapObject(p)
	if err != nil {
		t.Fatal(err)
	}
	ds.Set("p", obj)

	res, err := EvalString(`((get p "Double"))`, ds)
	if err != nil {
		t.Fatal(err)
	}
	i, ok := res.(IntAtom)
	if !ok || i.Value != 40 {
		t.Fatalf("expected IntAtom 40, got %T %v", res, res)
	}
}
