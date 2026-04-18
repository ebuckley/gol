---
name: gol-feature
description: Full feature development workflow for the gol interpreter. Use when the user wants to add a new language feature, Go interop capability, stdlib binding, or interpreter improvement. Covers planning, prereq analysis, implementation, code review, test coverage, documentation, and session close.
user-invocable: true
allowed-tools:
  - Read
  - Edit
  - Write
  - Glob
  - Grep
  - Bash
  - Agent
---

# /gol-feature — Feature Development Workflow

A structured workflow for adding new capabilities to the gol interpreter. This skill captures the full session pattern: plan → implement → review → test → document → ship.

Arguments: `$ARGUMENTS` — a short description of the feature to add, or empty to start a planning conversation.

---

## Phase 1: Planning

Before writing any code, understand what's needed and what it depends on.

### 1a. Understand the feature

If `$ARGUMENTS` is provided, use it as the starting point. Otherwise ask:
- What new behaviour should the interpreter/scripting engine have?
- Is this a language feature (new special form, new type), a Go interop improvement, or a stdlib binding?

### 1b. Identify prerequisites

Read the relevant source files to understand the current state:
- `lisp/eval.go` — special forms, evaluation logic
- `lisp/ast.go` — node types, Scope
- `lisp/object.go` — reflection layer, GoFunc, nodeToReflect
- `lisp/registry.go` — package registry and import
- `cmd/genpackage/main.go` — binding generator

Ask: **"Does anything need to exist before this can work?"** Common blockers in this codebase:
- A new type needs `goValueToNode` + `nodeToReflect` support before functions using it are bindable
- A new special form may depend on a runtime type not yet in the interpreter
- A stdlib import depends on the registry and namespace scopes being present

If prereqs are missing, list them as separate issues and order them before the main feature. Do not skip this — the session that built the scripting engine found 5 blocking prereqs before implementing `(import pkg)`.

### 1c. Create beads issues

Create a beads issue for each piece of work **before writing any code**:

```bash
bd create --title="<summary>" --description="<why this exists and what needs to be done>" --type=feature|task|bug --priority=1
```

For dependent work:
```bash
bd dep add <blocked-issue> <blocking-issue>
```

Check what's ready to start:
```bash
bd ready
```

---

## Phase 2: Implementation

Work through issues in dependency order. For each issue:

1. **Claim it**: `bd update <id> --claim`

2. **Write the code**. Key patterns in this codebase:
   - New node types go in `lisp/ast.go` and need `TokenLiteral()`, `IntLiteral()`, `FloatLiteral()` implementations
   - New special forms go in the `if/else if` chain in `lisp/eval.go`
   - New reflection support (new Go types) goes in `nodeToReflect` and `goValueToNode` in `lisp/object.go`
   - New package-level constants use the existing `nilNode`, `errType`, `anyType`, `nodeType` vars — never construct `BoolAtom{Atom: Atom{Token: Token{Type: SYMBOL, Literal: "nil"}}}` inline
   - Shared call-arg coercion uses `buildCallArgs` — don't duplicate the variadic loop

3. **Run tests after every change**:
   ```bash
   go test ./...
   ```

4. **Close the issue when done**: `bd close <id>`

5. **Commit each logical unit**:
   ```bash
   git add <files> && git commit -m "..."
   ```

---

## Phase 3: Code Review

After implementation, invoke the simplify skill to audit the changed code:

```
/simplify
```

Wait for all three review agents (reuse, quality, efficiency) to complete. Fix every real finding:
- Duplicated logic → extract a helper
- Inline construction of known constants → use the package var
- Hot-path allocations → cache at package level
- Leaky abstractions → fix the boundary

Do not skip this phase. In the session that built this scripting engine, the review found a duplicated variadic loop, 6 inline `nilNode` constructions, and hot-path `reflect.TypeOf` calls that should have been cached.

---

## Phase 4: Test Coverage

Explicitly audit coverage — do not rely on "tests pass" as a proxy for "tests are complete".

For each new piece of behaviour, verify these cases exist:

| Case | Check |
|------|-------|
| Happy path | Does the feature work with valid input? |
| Error path | Does it return a useful error for bad input? |
| Edge cases | Zero args, nil values, type mismatches, empty slices |
| Integration | Does it compose with existing features? |

Common gaps found in this codebase:
- Error paths in destructuring `:=` (fewer values than names → nil padding)
- Variadic with zero variadic args
- Import of unknown package
- Slice element values (not just length)

Add any missing tests and re-run:
```bash
go test ./... -v
```

---

## Phase 5: Documentation

If the change affects the **public API** (`lisp.*`), the **language** (new special form, new type, new error semantics), or **tooling** (`genpackage`, `stdlib/*`), update `README.md`.

Sections to consider:
- **Types table** — new runtime types
- **Special forms table** — new syntax
- **Error handling** — changes to error semantics
- **Go interop** — new `GoFunc` capabilities, new helpers
- **Importing packages** — new stdlib bindings
- **Embedding API** — new exported functions

Keep examples runnable — prefer `go run ./cmd/demo-*` over inline snippets where possible.

---

## Phase 6: Session Close

**Work is not complete until pushed.** Run this checklist in order:

```bash
go test ./...                          # all green
git status                             # see what changed
git add <files>                        # stage code changes
git commit -m "..."                    # commit
bd dolt push                           # push beads
git push                               # push to remote
git status                             # must show "up to date"
```

Commit message format:
```
<verb> <what>: <one line summary>

<bullet points of key decisions or non-obvious choices>

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>
```

---

## Reference: Codebase Conventions

**Node construction** — always use package vars, never inline:
```go
nilNode   // BoolAtom with literal "nil"
errType   // reflect.Type for error interface
anyType   // reflect.Type for any/interface{}
nodeType  // reflect.Type for Node interface
```

**Adding a new special form** — eval.go pattern:
```go
} else if fst.TokenLiteral() == "myform" {
    // validate L.Nodes length
    // eval sub-expressions as needed
    // return result, nil or nil, error
```

**Adding Go type support** — two places in object.go:
1. `nodeToReflect` — `case reflect.MyKind:` to convert Node → reflect.Value
2. `goValueToNode` — `case reflect.MyKind:` to convert reflect.Value → Node

**Generating stdlib bindings**:
```bash
go run ./cmd/genpackage -out stdlib <pkg> [pkg...]
```

Then blank-import the generated package and it auto-registers via `init()`.
