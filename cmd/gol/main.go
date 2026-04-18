// gol is the unified entry point for the gol interpreter.
// With no arguments it starts an interactive REPL; with a file argument it
// executes the script. All generated stdlib packages are pre-registered so
// scripts can use (import strings), (import fmt), etc.
package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/ebuckley/gol/lisp"
	_ "github.com/ebuckley/gol/stdlib/fmt"
	_ "github.com/ebuckley/gol/stdlib/strings"
)

const prompt = "gol> "

func main() {
	scope := lisp.DefaultScope()

	if len(os.Args) >= 2 {
		runFile(os.Args[1], scope)
	} else {
		runREPL(scope)
	}
}

func runFile(path string, scope *lisp.Scope) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gol: %v\n", err)
		os.Exit(1)
	}
	if _, err := lisp.EvalProgram(string(data), scope); err != nil {
		fmt.Fprintf(os.Stderr, "gol: %v\n", err)
		os.Exit(1)
	}
}

func runREPL(scope *lisp.Scope) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(prompt)
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		if line == "" {
			continue
		}
		result, err := lisp.EvalString(line, scope)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			continue
		}
		fmt.Println(result.TokenLiteral())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "gol: %v\n", err)
		os.Exit(1)
	}
}
