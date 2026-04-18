package lisp


type Token struct {
	Type TokenType
	Literal string
}

type TokenType string

const (
	SYMBOL = "SYMBOL"
	STRING = "STRING"
	LPAREN = "LPAREN"
	RPAREN = "RPAREN"
	ESCAPEFORM = "ESCAPE"
	COMMENTFORM = ";"
	EOF = "EOF"
)