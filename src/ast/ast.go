package ast

import (
  "JFFMonkeyLang/src/token"
)

// The base Node interface
type Node interface {
  TokenLiteral() string
  // String() string
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

/* Expressions */
type Identifier struct {
  Token token.Token // the token.IDENT token
  Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
