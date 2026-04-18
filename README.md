# GOL — Go Oriented Lisp

A Go-embedded Lisp interpreter and standalone scripting engine. Embed a scripting layer in any Go program, or run `.gol` scripts that import Go standard library packages directly.

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

| Type | Example | Notes |
|------|---------|-------|
| Integer | `42` | `int64` |
| Float | `3.14` | `float64` |
| Bool | `true` `false` | |
| String | `"hello"` | |
| Symbol | `foo` | |
| List | `(quote (1 2 3))` | also produced by Go slice returns |

### Special forms

| Form | Syntax |
|------|--------|
| Assignment | `(:= name expr)` |
| Destructuring | `(:= (a b) expr)` |
| Mutation | `(set! name expr)` |
| Conditional | `(if cond then else?)` |
| Sequence | `(do expr...)` |
| Lambda | `(func (params...) body)` |
| Quote | `(quote expr)` |
| Import | `(import pkg)` |

### Built-in functions

`+` `-` `*` `<` `>` `=` `println` `get`

### Error handling

Go functions that return `(T, error)` surface both values as a two-element list. Use destructuring `:=` to handle errors inline — no exceptions, just values:

```lisp
(:= (body err) (fetch "https://example.com"))
(if err
  (println err)
  (println body))
```

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
; stdlib strings via import
(import strings)
(:= greeting "hello, world")
(println (strings/to-upper greeting))
(println (strings/has-prefix greeting "hello"))
(:= parts (strings/split greeting ", "))
```

## Importing Go packages

GOL ships with generated bindings for Go standard library packages. Blank-import a package to register it, then use `(import name)` inside the script:

```go
import (
    "github.com/ebuckley/gol/lisp"
    _ "github.com/ebuckley/gol/stdlib/strings"
    _ "github.com/ebuckley/gol/stdlib/fmt"
)

scope := lisp.DefaultScope()
lisp.EvalProgram(`
  (import strings)
  (import fmt)
  (:= (result err) (fmt/errorf "oops: %s" "bad input"))
  (println err)
`, scope)
```

Namespace syntax `pkg/fn` calls a function inside an imported package:

```lisp
(strings/contains "hello" "ell")   ; => true
(strings/split "a,b,c" ",")        ; => ("a" "b" "c")
(strings/join (quote ("x" "y")) "-") ; => "x-y"
```

### Available packages

| Import path | Registered as |
|-------------|---------------|
| `github.com/ebuckley/gol/stdlib/strings` | `strings` |
| `github.com/ebuckley/gol/stdlib/fmt` | `fmt` |

## Generating bindings with genpackage

`cmd/genpackage` generates bindings for any installed Go package:

```bash
go run ./cmd/genpackage -out stdlib strings fmt math os
```

Each package produces a `stdlib/<pkg>/<pkg>_gol.go` file. The generated file:
- Converts Go names to lisp kebab-case (`HasPrefix` → `has-prefix`)
- Skips functions with unsupported types (channels, maps, unsafe pointers)
- Registers via `init()` so a blank import activates it automatically

For packages outside your module, generate into your own directory:

```bash
go run github.com/ebuckley/gol/cmd/genpackage -out ./mypkg/gol github.com/some/library
```

## Go interop

### Binding Go functions

`GoFunc` wraps any Go function as a Lisp callable, including variadic functions. Arguments and return values are automatically coerced between Lisp nodes and Go types.

```go
scope.Set("upper",    lisp.GoFunc(strings.ToUpper))
scope.Set("contains", lisp.GoFunc(strings.Contains))
scope.Set("sprintf",  lisp.GoFunc(fmt.Sprintf))   // variadic
scope.Set("split",    lisp.GoFunc(strings.Split))  // returns []string → List
```

```lisp
(upper "hello")                    ; => "HELLO"
(contains "hello" "ell")           ; => true
(sprintf "%s=%d" "x" 42)          ; => "x=42"
(split "a,b,c" ",")               ; => ("a" "b" "c")
```

Go slices are automatically converted to Lisp lists and back, so slice-returning functions compose naturally:

```lisp
(join (split "a,b,c" ",") "-")    ; => "a-b-c"
```

### Namespace scopes

Register a whole package as a namespace so its symbols don't pollute the top-level scope:

```go
scope.SetNamespace("str", map[string]lisp.Node{
    "contains": lisp.GoFunc(strings.Contains),
    "upper":    lisp.GoFunc(strings.ToUpper),
})
```

```lisp
(str/contains "hello" "ell")   ; => true
(str/upper "hello")            ; => "HELLO"
```

### Wrapping Go structs

`WrapObject` uses reflection to expose a struct's exported fields and methods. Access them with the built-in `get`.

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
(get resp "Status")      ; => "200 OK"
(get resp "StatusCode")  ; => 200
```

### Injecting opaque Go values

For values you want to pass through the interpreter without introspection:

```go
scope.Bind("db", myDB)           // stores as GoValue
val, ok := lisp.Unwrap(node)     // extract back to any
```

### Manual node conversion

```go
node := lisp.ToNode("hello")            // any → Node
str, err := lisp.FromNode[string](node) // Node → typed Go value
```

## Package registry

To make a package importable via `(import name)` at script runtime, register it:

```go
// In your package's init() or setup:
lisp.Register("mylib", func(scope *lisp.Scope) {
    scope.Set("do-thing", lisp.GoFunc(mylib.DoThing))
    scope.Set("other",    lisp.GoFunc(mylib.Other))
})

// Or load it programmatically:
lisp.ImportInto("mylib", scope)
```

## Embedding API

```go
// Evaluate a single expression
lisp.EvalString(code string, scope *Scope) (Node, error)

// Evaluate multiple top-level forms, return last result
lisp.EvalProgram(code string, scope *Scope) (Node, error)

// Create a scope with built-in functions pre-loaded
lisp.DefaultScope() *Scope

// Register a package so (import name) works
lisp.Register(name string, fn RegisterFunc)

// Load a registered package into a scope directly
lisp.ImportInto(name string, scope *Scope) error

// Register a map of functions under a namespace
scope.SetNamespace(ns string, fns map[string]Node)

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
| `go run ./cmd/demo-strings` | Generated `stdlib/strings` binding via `(import strings)` |
| `go run ./cmd/demo-store` | Wrapping a mutable `*Store` struct with `WrapObject` |
| `CGO_ENABLED=0 go run ./cmd/demo-http` | HTTP client with Go-style `(:= (val err) ...)` error handling |
| `go run ./cmd/repl` | Interactive REPL |
| `go run ./cmd/interp <file.gol>` | Run a `.gol` file |

Lisp example programs are in `_examples/`.

## Inspired by

- https://github.com/bytbox/kakapo
- http://norvig.com/lispy.html
