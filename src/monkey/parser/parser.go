package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}
	// prefixParseFnsの初期化
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	// 構文解析関数をprefixParseFnsに登録
	// token.IDENTが出現したらp.parseIdentifierが呼ばれる？
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)

	// 2つのトークンを読み込む。curTokenとpeekTokenの両方がセットされる
	//p.curToken = nil p.peekToken = 1つ目のトークン
	p.nextToken()
	//p.curToken = 1つ目のトークン p.peekToken = 2つ目のトークン
	p.nextToken()

	return p
}

func (p *Parser) parseIdentifier() ast.Expression {
	// ast.Expressionのほうがast.Identifierより抽象度が高い
	// ast.Identifierはast.Expressionのインターフェイスを実装している
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
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
			// let文が3つ入る想定
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
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	// 現在はletなので次はIDENTが必ず来るはずなのでチェックする
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	//上のチェックを通っていれば１つ次のトークンに進んでいるので現在の識別子トークンをセットする
	// x, y、foobar とか
	//LetStatementのName(識別子)にast.Identifierを入れておく
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// let x と来たら次は=が来るはずなのでチェックする
	if !p.expectPeek(token.ASSIGN) {
		return nil
	}
	// Todo: セミコロンに遭遇するまで読みとばしてしまっている
	// curTokenがセミコロンでない場合次のトークンに進む
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s insted",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}
func (p *Parser) parseReturnStatement() ast.Statement {
	// 現在のトークンは'return'トークン
	stmt := &ast.ReturnStatement{Token: p.curToken}

	// curTokeが式(5,10,993322)になり、nextTokenが;になる
	p.nextToken()

	// TODO: セミコロンに遭遇するまで読み飛ばしてしまっている
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	// p.parseIdentifierが呼ばれる
	// ast.Expressionが返ってくる
	// 前置演算子の場合、leftという名前がよくわからない（2018-12-12）
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		// ここで現在のトークンが15から+に変わる
		p.nextToken()
		// infixはparseInfixExpression()とか
		leftExp = infix(leftExp)
	}

	return leftExp
}

// ここで優先順位が決まる？
const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > または <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X または !X
	CALL        // myFunction(X)
)

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value

	return lit
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	// この関数内で!15みたいな式をすべて解析してしまう
	// curTokenは15に設定される
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

// precedenceは「順位」という意味
// 下に行くに連れ優先順位が上がる
var precedence = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
}

// 次のトークンタイプの優先順位のナンバーを返す
func (p *Parser) peekPrecedence() int {
	if p, ok := precedence[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// 現在んのトークンタイプの優先順位のナンバーを返す
func (p *Parser) curPrecedence() int {
	if p, ok := precedence[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// 現在のトークンが中間演算子の場合parseExpression()から呼び出される
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left, // 15 + 13 の15のastが格納される
	}
	// + 等の中間演算子の優先順位が格納される
	precedence := p.curPrecedence()
	// 15 + 13 の場合現在の位置が13になる
	p.nextToken()
	// 13 のastが返る
	expression.Right = p.parseExpression(precedence)

	return expression
}
