package main

import (
	"bufio"
	"ersin.nz/gol/lisp"
	"fmt"
	"os"
)

const PROMPT = "gol> "

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for  {
		fmt.Print(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			break
		}
		line := scanner.Text()
		l := lisp.NewLexer(line)
		node, err := lisp.NewASTFromLex(l)
		if err != nil {
			fmt.Printf("Parsing Error:\n%s\n", err.Error())
			continue
		}
		r, err := lisp.Eval(node, lisp.DefaultScope())
		if err != nil {
			fmt.Printf("Evaluation Error:\n%s\n", err.Error())
			continue
		}
		fmt.Print(r.TokenLiteral(), "\n")
	}
}

