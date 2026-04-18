// demo-strings shows how to bind Go's strings package into a gol scope.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ebuckley/gol/lisp"
)

const program = `
(do
  (:= greeting "hello, world")

  (println (upper greeting))
  (println (title "go embedded lisp"))

  (println (contains greeting "world"))
  (println (contains greeting "rust"))

  (:= swapped (replace greeting "world" "gol"))
  (println swapped)

  (println (has-prefix greeting "hello"))
  (println (has-suffix greeting "world"))

  (println (trim-space "   lots of whitespace   ")))
`

func main() {
	scope := lisp.DefaultScope()

	scope.Set("upper", lisp.GoFunc(strings.ToUpper))
	scope.Set("lower", lisp.GoFunc(strings.ToLower))
	scope.Set("title", lisp.GoFunc(strings.ToTitle))
	scope.Set("contains", lisp.GoFunc(strings.Contains))
	scope.Set("replace", lisp.GoFunc(func(s, old, new string) string {
		return strings.ReplaceAll(s, old, new)
	}))
	scope.Set("has-prefix", lisp.GoFunc(strings.HasPrefix))
	scope.Set("has-suffix", lisp.GoFunc(strings.HasSuffix))
	scope.Set("trim-space", lisp.GoFunc(strings.TrimSpace))

	_, err := lisp.EvalProgram(program, scope)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
