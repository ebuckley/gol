package main

import (
	"bufio"
	"github.com/ebuckley/gol"
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

		r, err := lisp.EvalString(line, lisp.DefaultScope())
		if err != nil {
			fmt.Printf("Evaluation Error:\n%s\n", err.Error())
			continue
		}
		fmt.Print(r.TokenLiteral(), "\n")
	}
}

