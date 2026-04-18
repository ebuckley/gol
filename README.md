# GOL — Go Oriented Lisp

A Go-embedded Lisp interpreter. Embed a scripting layer in any Go program with full two-way interop: call Go functions from Lisp, wrap Go structs and expose their methods, and pass native Go values through the interpreter.

## Quick start

```go
import "github.com/ebuckley/gol/lisp"

scope := lisp.DefaultScope()
result, err := lisp.EvalString(`(+ 1 2)`, scope)
```

Evaluate a multi-expression program:

```go
result, err := lisp.EvalProgram(`
  (:= square (func (x) (* x x)))
  (square 9)
`, scope)
```

## Language

### Types

| Type | Example |
|------|---------|
| Integer | `42` |
| Float | `3.14` |
| Bool | `true` `false` |
| String | `"hello"` |
| Symbol | `foo` |
| List | `(+ 1 2)` |

### Special forms

| Form | Syntax |
|------|--------|
| Assignment | `(:= name expr)` |
| Mutation | `(set! name expr)` |
| Conditional | `(if cond then else?)` |
| Sequence | `(do expr...)` |
| Lambda | `(func (params...) body)` |
| Quote | `(quote expr)` |

### Built-in functions

`+` `-` `*` `<` `>` `=` `println` `get`

### Example programs

```lisp
; factorial
(:= factorial (func (n)
  (if (< n 2)
    1
    (* n (factorial (- n 1))))))

(println (factorial 10))
```

```lisp
; string operations (when bound by the host)
(:= greeting "hello, world")
(println (upper greeting))
(println (contains greeting "world"))
```

## Go interop

### Binding Go functions

`GoFunc` wraps any Go function as a Lisp callable. Arguments and return values are automatically coerced between Lisp nodes and Go types (`string`, `int64`, `float64`, `bool`, `error`).

```go
scope.Set("upper",    lisp.GoFunc(strings.ToUpper))
scope.Set("contains", lisp.GoFunc(strings.Contains))
scope.Set("sqrt",     lisp.GoFunc(math.Sqrt))
```

```lisp
(upper "hello")          ; => "HELLO"
(contains "hello" "ell") ; => true
(sqrt 144)               ; => 12
```

### Wrapping Go structs

`WrapObject` uses reflection to expose a struct's exported fields and methods as an `ObjectNode`. Access them with the built-in `get`.

```go
type Store struct { ... }
func (s *Store) Set(key, value string) { ... }
func (s *Store) Get(key string) string  { ... }

obj, _ := lisp.WrapObject(myStore)
scope.Set("store", obj)
```

```lisp
(:= store-set (get store "Set"))
(:= store-get (get store "Get"))

(store-set "name" "Alice")
(println (store-get "name"))   ; => Alice
```

Exported struct fields are also accessible:

```lisp
(get resp "Status")     ; => "200 OK"
(get resp "StatusCode") ; => 200
```

### Injecting opaque Go values

For values you want to pass through the interpreter without introspection:

```go
scope.Bind("db", myDB)           // stores as GoValue
val, ok := lisp.Unwrap(node)     // extract back to any
```

### Manual node conversion

```go
node := lisp.ToNode("hello")     // any → Node
str, err := lisp.FromNode[string](node)  // Node → typed Go value
```

## Embedding API

```go
// Evaluate a single expression
lisp.EvalString(code string, scope *Scope) (Node, error)

// Evaluate multiple top-level forms, return last result
lisp.EvalProgram(code string, scope *Scope) (Node, error)

// Create a scope with built-in functions pre-loaded
lisp.DefaultScope() *Scope

// Wrap/unwrap arbitrary Go values
lisp.Wrap(v any) Node
lisp.Unwrap(n Node) (any, bool)

// Wrap a struct with field/method access
lisp.WrapObject(v any) (ObjectNode, error)

// Wrap a Go function as a Lisp callable
lisp.GoFunc(fn any) Callable

// Convert between Go values and Nodes
lisp.ToNode(v any) Node
lisp.FromNode[T any](n Node) (T, error)
```

## Example programs

| Program | What it shows |
|---------|---------------|
| `go run ./cmd/demo-strings` | Binding `strings` stdlib functions via `GoFunc` |
| `go run ./cmd/demo-store` | Wrapping a mutable `*Store` struct with `WrapObject` |
| `CGO_ENABLED=0 go run ./cmd/demo-http` | HTTP client — wrapping `*http.Response` fields |
| `go run ./cmd/repl` | Interactive REPL |
| `go run ./cmd/interp <file.gol>` | Run a `.gol` file |

Lisp example programs are in `_examples/`.

## Inspired by

- https://github.com/bytbox/kakapo
- http://norvig.com/lispy.html
