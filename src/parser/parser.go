package parser

import (
  "JFFMonkeyLang/src/ast"
  "JFFMonkeyLang/src/lexer"
  "JFFMonkeyLang/src/token"
  "fmt"
  "strconv"
)

const (
  _ int = iota
  LOWEST
  EQUALS      // ==
  LESSGREATER // > or <
  SUM         // +
  PRODUCT     // *
  PREFIX      // -X or !X
  CALL        // myFunction(X)
)

var precedences = map[token.TokenType]int{
  token.EQ:       EQUALS,
  token.NOT_EQ:   EQUALS,
  token.LT:       LESSGREATER,
  token.GT:       LESSGREATER,
  token.PLUS:     SUM,
  token.MINUS:    SUM,
  token.SLASH:    PRODUCT,
  token.ASTERISK: PRODUCT,
  token.LPAREN:   CALL,
}

type (
  prefixParseFn func() ast.Expression
  infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
  l      *lexer.Lexer
  errors []string

  curToken  token.Token
  peekToken token.Token

  //           ┌-> prefixParseFn
  // curToken ─┤
  //           └-> infixParseFn
  prefixParseFns map[token.TokenType]prefixParseFn
  infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
  p := &Parser{l: l}

  p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
  p.registerPrefix(token.IDENT, p.parseIdentifier)       // eg: foo
  p.registerPrefix(token.INT, p.parseIntegerLiteral)     // eg: 5
  p.registerPrefix(token.BANG, p.parsePrefixExpression)  // eg: "!5"
  p.registerPrefix(token.MINUS, p.parsePrefixExpression) // eg: "-5"
  p.registerPrefix(token.TRUE, p.parseBoolean)           // eg: true
  p.registerPrefix(token.FALSE, p.parseBoolean)          // eg: false

  p.infixParseFns = make(map[token.TokenType]infixParseFn)
  p.registerInfix(token.PLUS, p.parseInfixExpression)     // 1 + 1
  p.registerInfix(token.MINUS, p.parseInfixExpression)    // 1 - 1
  p.registerInfix(token.SLASH, p.parseInfixExpression)    // 1 / 1
  p.registerInfix(token.ASTERISK, p.parseInfixExpression) // 1 + 1
  p.registerInfix(token.EQ, p.parseInfixExpression)       // 1 == 1
  p.registerInfix(token.NOT_EQ, p.parseInfixExpression)   // 1 != 1
  p.registerInfix(token.LT, p.parseInfixExpression)       // 1 < 1
  p.registerInfix(token.GT, p.parseInfixExpression)       // 1 > 1

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
    return p.parseExpressionStatement()
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

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
  // debug print
  defer untrace(trace("parseExpressionStatement"))

  // 1.build AST node
  stmt := &ast.ExpressionStatement{Token: p.curToken}
  // 2.defalut precedence is LOWEST
  stmt.Expression = p.parseExpression(LOWEST)

  // 3.check ';' and jump ';' token
  if p.peekTokenIs(token.SEMICOLON) {
    p.nextToken()
  }

  return stmt
}

// !!! Pratt Parsing Core Logic !!!
func (p *Parser) parseExpression(precedence int) ast.Expression {
  // debug print
  defer untrace(trace("parseExpression"))

  prefixFn := p.prefixParseFns[p.curToken.Type]

  if prefixFn == nil {
    p.noPrefixParseFnError(p.curToken.Type)
    return nil
  }
  leftExpression := prefixFn()

  for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
    infixFn := p.infixParseFns[p.peekToken.Type]
    if infixFn == nil {
      return leftExpression
    }

    p.nextToken()

    leftExpression = infixFn(leftExpression)
  }

  return leftExpression
}

// eg: foo;
func (p *Parser) parseIdentifier() ast.Expression {
  return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBoolean() ast.Expression {
  return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

// eg: 5;
func (p *Parser) parseIntegerLiteral() ast.Expression {
  // debug print
  defer untrace(trace("parseIntegerLiteral " + p.curToken.Literal))

  literal := &ast.IntegerLiteral{Token: p.curToken}

  // string to int
  value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
  if err != nil {
    msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
    p.errors = append(p.errors, msg)
    return nil
  }

  literal.Value = value

  return literal
}

// eg: !5, -5
func (p *Parser) parsePrefixExpression() ast.Expression {
  // debug print
  defer untrace(trace("parsePrefixExpression " + p.curToken.Literal))

  expression := &ast.PrefixExpression{
    Token:    p.curToken,
    Operator: p.curToken.Literal,
  }

  // jump Prefix Token
  p.nextToken()

  // recursive parsing
  expression.Right = p.parseExpression(PREFIX)

  return expression
}

// eg: 1 + 2
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
  // debug print
  defer untrace(trace("parseInfixExpression " + p.curToken.Literal))

  expression := &ast.InfixExpression{
    Token:    p.curToken,         // eg: +
    Operator: p.curToken.Literal, // eg: +
    Left:     left,               // eg: 1
  }

  precedence := p.curPrecedence() // eg: SUM(4)

  p.nextToken()

  // recursive parsing
  expression.Right = p.parseExpression(precedence)

  return expression
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

func (p *Parser) peekPrecedence() int {
  if p, ok := precedences[p.peekToken.Type]; ok {
    return p
  }

  return LOWEST
}

func (p *Parser) curPrecedence() int {
  if p, ok := precedences[p.curToken.Type]; ok {
    return p
  }

  return LOWEST
}

func (p *Parser) Errors() []string {
  return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
  msg := fmt.Sprintf("expected next token to be %s, got %s instead",
    t, p.peekToken.Type)
  p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
  msg := fmt.Sprintf("no prefix parse function for %s found", t)
  p.errors = append(p.errors, msg)
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
  p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
  p.infixParseFns[tokenType] = fn
}
