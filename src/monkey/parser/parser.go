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
	program := &ast.Program{}
	program.Statements = []ast.Statement{} // ast.Statement型のスライスを格納

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			// program.Statementsにstmtを入れて、それを返す(毎回上書き？)
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		// *ast.LetStatementはast.Statementのインターフェイスを満たしているためast.Statement型となる
		return p.parseLetStatement()
	default:
		return nil
	}
}
func (p *Parser) parseLetStatement() *ast.LetStatement {

}
