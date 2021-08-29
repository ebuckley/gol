# GOL is: Go oriented LISP


```go
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
```


Inspired by https://github.com/bytbox/kakapo

Goals:

- Have fun
- Create a lisp that can be deployed anywhere.
- Provide a familiar experience for programmers coming from go.
- Provide easy interop with the go language.



# Roadmap
- improvements to lexer/parser so that it doesn't break as often.
- scan go package and convert to environment functions
- more essential builtins
- better REPL (move up/down)
- integration with nrepl protocol