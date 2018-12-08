package ast

import "monkey/token"

type Node interface {
	TokenLiteral() string
}

// 文を表現する(let x = 5 みたいなやつ)
type Statement interface {
	Node
	statementNode()
}

//式を表現する(5 とか)
type Expression interface {
	Node
	expressionNode()
}

// 文のスライス格納される
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

// let文全体
type LetStatement struct {
	Token token.Token
	Name  *Identifier // 識別子
	Value Expression  // 式
}

func (ls *LetStatement) statementNode() {}

func (ls *LetStatement) TokenLiteral() string {
	return ls.Token.Literal
}

// 識別子
type Identifier struct {
	Token token.Token // Token.IDENT トークン
	Value string
}

func (i *Identifier) expressionNode() {}

func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

type ReturnStatement struct {
	Token       token.Token // 'returnトークン'
	returnValue Expression
}

func (rs *ReturnStatement) statementNode() {}

func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}
