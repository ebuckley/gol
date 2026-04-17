# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run all tests
go test ./...

# Run tests in a specific package
go test ./lisp/...

# Run a single test
go test ./lisp/... -run TestSimple

# Build the REPL
go build ./cmd/repl

# Run the REPL
go run ./cmd/replmodule "github.com/ebuckley/gol"
```

## Architecture

`gol` is a Go-embedded Lisp interpreter. The core pipeline is: **lex → parse (AST) → eval**.

All interpreter logic lives in the `lisp` package:

- **`lex.go`** — tokenizes input into `Token` values (types defined in `token.go`)
- **`ast.go`** — `NewASTFromLex` builds a tree of `Node` values; also defines the type hierarchy (`List`, `Atom`, `SymbolAtom`, `IntAtom`, `FloatAtom`, `BoolAtom`, `Callable`) and the `Scope` chain (lexically scoped env with parent pointer)
- **`eval.go`** — `Eval(node, scope)` walks the AST; special forms (`if`, `:=`, `do`, `func`, `quote`, `set!`) are handled inline; function application evaluates all args then calls the `Callable`; `DefaultScope()` seeds the env with built-in functions (`+`, `*`)
- **`parse.go`** — thin wrapper (likely `NewASTFromLex` delegation)

The public API for embedding is just two functions:
```go
lisp.EvalString(code string, scope *Scope) (Node, error)
lisp.DefaultScope() *Scope
```

`cmd/repl` is a thin stdin loop over `EvalString`. `cmd/genpackage` is a stub intended to auto-generate Go→gol bindings (not yet implemented).

## Special forms

| Form | Syntax |
|------|--------|
| assignment | `(:= name expr)` |
| conditional | `(if cond then else?)` |
| sequence | `(do expr...)` |
| lambda | `(func (params...) body)` |
| mutation | `(set! name expr)` |
| quote | `(quote expr)` |
