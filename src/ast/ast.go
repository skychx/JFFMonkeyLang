package ast

import (
  "JFFMonkeyLang/src/token"
  "bytes"
)

// The base Node interface
type Node interface {
  TokenLiteral() string
  String() string // print ast node for test
}

// All statement nodes implement this
type Statement interface {
  Node
  statementNode()
}

// All expression nodes implement this
type Expression interface {
  Node
  expressionNode()
}

// ast root node
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

func (p *Program) String() string {
  var out bytes.Buffer

  for _, s := range p.Statements {
    out.WriteString(s.String())
  }

  return out.String()
}

/* Statements */

/*
 * let   a       = 5
 * Token Name      Value
 * Token Identifier Expression
 */
type LetStatement struct {
  Token token.Token // the token.LET token
  Name  *Identifier
  Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
  var out bytes.Buffer

  out.WriteString(ls.TokenLiteral() + " ")
  out.WriteString(ls.Name.String())
  out.WriteString(" = ")

  if ls.Value != nil {
    out.WriteString(ls.Value.String())
  }

  out.WriteString(";")

  return out.String()
}

/*
 * return 5;
 * return add(1, 2)
 * Token  ReturnValue
 */
type ReturnStatement struct {
  Token       token.Token
  ReturnValue Expression
}

func (ls *ReturnStatement) statementNode()       {}
func (ls *ReturnStatement) TokenLiteral() string { return ls.Token.Literal }
func (rs *ReturnStatement) String() string {
  var out bytes.Buffer

  out.WriteString(rs.TokenLiteral() + " ")

  if rs.ReturnValue != nil {
    out.WriteString(rs.ReturnValue.String())
  }

  out.WriteString(";")

  return out.String()
}

/*
 * x     x + 10;
 * Token Expression
 *
 *
 */
type ExpressionStatement struct {
  Token      token.Token // the first token of the expression
  Expression Expression
}

func (ls *ExpressionStatement) statementNode()       {}
func (ls *ExpressionStatement) TokenLiteral() string { return ls.Token.Literal }
func (es *ExpressionStatement) String() string {
  if es.Expression != nil {
    return es.Expression.String()
  }
  return ""
}

/* Expressions */
type Identifier struct {
  Token token.Token // the token.IDENT token
  Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// eg: {Token: token.INT, Value: 5}
type IntegerLiteral struct {
  Token token.Token
  Value int64 // monkeyLang number only has int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// eg: !5, -5
type PrefixExpression struct {
  Token    token.Token // The prefix token, e.g: !
  Operator string      // "-" or "!"
  Right    Expression  // eg: 5
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
  var out bytes.Buffer

  out.WriteString("(")
  out.WriteString(pe.Operator)
  out.WriteString(pe.Right.String())
  out.WriteString(")")

  return out.String()
}

// eg: 1 + 1
type InfixExpression struct {
  Token    token.Token // The infix token, e.g: +
  Left     Expression  // eg: 1
  Operator string      // "+"
  Right    Expression  // eg: 1
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
  var out bytes.Buffer

  out.WriteString("(")
  out.WriteString(ie.Left.String())
  out.WriteString(" " + ie.Operator + " ")
  out.WriteString(ie.Right.String())
  out.WriteString(")")

  return out.String()
}
