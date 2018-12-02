package parser

import (
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
)

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peekToken token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	// 2つのトークンを読み込む。curTokenとpeekTokenの両方がセットされる
	//p.curToken = nil p.peekToken = 1つ目のトークン
	p.nextToken()
	//p.curToken = 1つ目のトークン p.peekToken = 2つ目のトークン
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	return nil
}
