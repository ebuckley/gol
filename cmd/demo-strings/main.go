// demo-strings shows the generated stdlib/strings binding used via (import strings).
package main

import (
	"fmt"
	"os"

	"github.com/ebuckley/gol/lisp"
	_ "github.com/ebuckley/gol/stdlib/strings"
)

const program = `
(do
  (import strings)
  (:= greeting "hello, world")

  (println (strings/to-upper greeting))
  (println (strings/contains greeting "world"))
  (println (strings/contains greeting "rust"))

  (:= swapped (strings/replace-all greeting "world" "gol"))
  (println swapped)

  (println (strings/has-prefix greeting "hello"))
  (println (strings/has-suffix greeting "world"))
  (println (strings/trim-space "   lots of whitespace   "))

  (:= parts (strings/split greeting ", "))
  (println parts))
`

func main() {
	scope := lisp.DefaultScope()
	_, err := lisp.EvalProgram(program, scope)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
