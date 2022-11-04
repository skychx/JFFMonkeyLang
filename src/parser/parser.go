package parser

import (
  "JFFMonkeyLang/src/ast"
  "JFFMonkeyLang/src/lexer"
  "JFFMonkeyLang/src/token"
  "fmt"
)

type Parser struct {
  l      *lexer.Lexer
  errors []string

  curToken  token.Token
  peekToken token.Token
}

func New(l *lexer.Lexer) *Parser {
  p := &Parser{l: l}

  // Read two tokens, so curToken and peekToken are both set
  p.nextToken()
  p.nextToken()

  return p
}

func (p *Parser) nextToken() {
  p.curToken = p.peekToken
  p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
  // 1. build ast root node
  program := &ast.Program{}
  program.Statements = []ast.Statement{}

  // 2. Parse the token continuously and identify the statement
  // until the end of the file
  for !p.curTokenIs(token.EOF) {
    stmt := p.parseStatement()
    if stmt != nil {
      program.Statements = append(program.Statements, stmt)
    }
    p.nextToken()
  }

  return program
}

/* parse Statements */
func (p *Parser) parseStatement() ast.Statement {
  switch p.curToken.Type {
  case token.LET:
    return p.parseLetStatement()
  case token.RETURN:
    return p.parseReturnStatement()
  default:
    return nil
  }
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
  // 1
  stmt := &ast.LetStatement{Token: p.curToken} // token.LET

  // 2.Check variable name(expect token). eg: x, y, foo, bar
  if !p.expectPeek(token.IDENT) {
    // When `nil`` is returned here,
    // ParseProgram will filter and skip the parsing of the statement,
    // which is equivalent to eating the Error,
    // a more robust way is to throw an error and terminate the parsing
    return nil
  }

  stmt.Name = &ast.Identifier{
    Token: p.curToken, // token.IDENT
    Value: p.curToken.Literal,
  }

  // 3.Check '=' token
  if !p.expectPeek(token.ASSIGN) {
    //
    return nil
  }

  // TODO: Temporarily skip the processing of expressions
  // until a semicolon is encountered
  for !p.curTokenIs(token.SEMICOLON) {
    p.nextToken()
  }

  return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
  // 1.
  stmt := &ast.ReturnStatement{Token: p.curToken} // token.RETURN

  p.nextToken()

  // TODO: Temporarily skip the processing of expressions
  // until a semicolon is encountered
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
  msg := fmt.Sprintf("expected next token to be %s, got %s instead",
    t, p.peekToken.Type)
  p.errors = append(p.errors, msg)
}
