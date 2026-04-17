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


<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:ca08a54f -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

## Session Completion

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd dolt push
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
<!-- END BEADS INTEGRATION -->
