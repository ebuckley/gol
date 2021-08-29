package lisp


type Parser struct {
	l *Lexer
	currToken Token
	peekToken Token
}

func (p *Parser) nextToken() {
	p.currToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

//TODO return the ast
func (p *Parser) ParseProgram() {

}


func NewParser(l *Lexer) *Parser {
	p := &Parser{l: l}
	p.nextToken()
	p.nextToken()
	return p
}