package lisp

import (
	"io"
	"strings"
	"unicode"
)

type Lexer struct {
	rr io.RuneScanner

	// position of the current position in the rune reading
	position int
	// ch references the current rune in the lexer
	ch rune
}

func (l *Lexer) readRune() {
	readRune, _, err := l.rr.ReadRune()
	if err == io.EOF {
		l.position += 1
		l.ch = 0
		return
	} else if err != nil {
		panic(err)
	}
	l.ch = readRune
	l.position += 1
}

func (l *Lexer) peekRune() rune {
	readRune, _, err := l.rr.ReadRune()
	if err == io.EOF {
		return 0
	} else if err != nil {
		panic(err)
	}
	err = l.rr.UnreadRune()
	if err != nil {
		panic(err)
	}
	return readRune
}

func shouldChomp(r rune) bool {
	chompable := unicode.IsSpace(r) || !unicode.IsPrint(r) || unicode.IsControl(r)
	return chompable && r != 0
}

// chompRunes will remove whitespace, control, non printable runes and anything else illegal
func (l *Lexer) chompRunes() {
	for shouldChomp(l.ch) {
		l.readRune()
	}
}

func (l *Lexer) readSymbol() string {
	// simply read until the next space or control rune
	var newSymbol string
	for !unicode.IsSpace(l.ch) && l.ch != '(' && l.ch != ')' && l.ch != ';' && l.ch !='\'' {
		newSymbol = newSymbol +  string(l.ch)
		l.readRune()
		if l.ch == 0 {
			break; // an early escape if we read to end of file while reading the current symbol (*
		}
	}
	return newSymbol
}

func (l *Lexer) NextToken() Token {
	var tok Token
	l.chompRunes()
	switch l.ch {
	case '(':
		tok = Token{
			Type:    LPAREN,
			Literal: "",
		}
	case ')':
		tok = Token{
			Type:    RPAREN,
			Literal: "",
		}
		//TODO case escapeform (for macros :o)
	//TODO case COMMENTFORM:
	// ignore the rest and go to the end of the line
	case 0:
		tok = Token{
			Type:    EOF,
			Literal: "",
		}
	default:
		// TODO: also read numbers, strings, atoms, keywords, builtins etc..

		tok = Token{
			Type:    SYMBOL,
			Literal: l.readSymbol(),
		}
		return tok
	}
	l.readRune() // advance the token reader by one
	return tok
}

func NewLexer(inp string) *Lexer {
	rr := io.RuneScanner(strings.NewReader(inp))

	l := &Lexer{rr: rr}
	l.readRune()
	return l
}
