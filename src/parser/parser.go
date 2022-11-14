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
  p.registerPrefix(token.IDENT, p.parseIdentifier)         // eg: foo
  p.registerPrefix(token.INT, p.parseIntegerLiteral)       // eg: 5
  p.registerPrefix(token.BANG, p.parsePrefixExpression)    // eg: "!5"
  p.registerPrefix(token.MINUS, p.parsePrefixExpression)   // eg: "-5"
  p.registerPrefix(token.TRUE, p.parseBoolean)             // eg: true
  p.registerPrefix(token.FALSE, p.parseBoolean)            // eg: false
  p.registerPrefix(token.LPAREN, p.parseGroupedExpression) // eg: (
  p.registerPrefix(token.IF, p.parseIfExpression)          // eg: if
  p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral) // eg: fn() { return foo; }

  p.infixParseFns = make(map[token.TokenType]infixParseFn)
  p.registerInfix(token.PLUS, p.parseInfixExpression)     // 1 + 1
  p.registerInfix(token.MINUS, p.parseInfixExpression)    // 1 - 1
  p.registerInfix(token.SLASH, p.parseInfixExpression)    // 1 / 1
  p.registerInfix(token.ASTERISK, p.parseInfixExpression) // 1 + 1
  p.registerInfix(token.EQ, p.parseInfixExpression)       // 1 == 1
  p.registerInfix(token.NOT_EQ, p.parseInfixExpression)   // 1 != 1
  p.registerInfix(token.LT, p.parseInfixExpression)       // 1 < 1
  p.registerInfix(token.GT, p.parseInfixExpression)       // 1 > 1

  p.registerInfix(token.LPAREN, p.parseCallExpression) // add(1, 2)

  // Read two tokens, so curToken and peekToken are both set
  p.nextToken()
  p.nextToken()

  return p
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
  stmt := &ast.LetStatement{Token: p.curToken} // token.LET

  // 1.curToken is 'let', peekToken may be IDENT
  // let a = 1;
  // ....^.....
  if !p.expectPeek(token.IDENT) {
    // When `nil` is returned here,
    // ParseProgram will filter and skip the parsing of the statement,
    // which is equivalent to eating the Error,
    // a more robust way is to throw an error and terminate the parsing
    return nil
  }

  // 2.curToken is IDENT
  stmt.Name = &ast.Identifier{
    Token: p.curToken, // token.IDENT
    Value: p.curToken.Literal,
  }

  // 3.curToken is IDENT, peekToken may be '='
  // let a = 1;
  // ......^...
  if !p.expectPeek(token.ASSIGN) {
    return nil
  }

  // 4.curToken is '=', jump it
  p.nextToken()

  // 5.parseExpression
  stmt.Value = p.parseExpression(LOWEST)

  // 6.peekToken may be ';'
  // let a = 1;
  // .........^
  if p.peekTokenIs(token.SEMICOLON) {
    // 7.peekToken is ';', jump to it
    p.nextToken()
  }
  // 8.curToken is ';'

  return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
  stmt := &ast.ReturnStatement{Token: p.curToken} // token.RETURN

  // 1.curToken is 'return', jump it
  p.nextToken()
  // 2.parseExpression
  stmt.ReturnValue = p.parseExpression(LOWEST)

  // 3.peekToken may be ';'
  // return a;
  // ........^
  for p.peekTokenIs(token.SEMICOLON) {
    // 4.peekToken is ';', jump to it
    p.nextToken()
  }
  // 5.curToken is ';'

  return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
  // debug print
  defer untrace(trace("parseExpressionStatement"))

  // 1.build AST node
  stmt := &ast.ExpressionStatement{Token: p.curToken}
  // 2.defalut precedence is LOWEST
  stmt.Expression = p.parseExpression(LOWEST)

  // 3.peekToken may be ';'
  if p.peekTokenIs(token.SEMICOLON) {
    // 4.peekToken is ';', jump to it
    p.nextToken()
  }
  // 5.curToken is ';'

  return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
  block := &ast.BlockStatement{Token: p.curToken}
  block.Statements = []ast.Statement{}

  // 1.curToken is '{', jump it
  // { a }
  // ^....
  p.nextToken()

  for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
    statement := p.parseStatement()
    if statement != nil {
      block.Statements = append(block.Statements, statement)
    }
    p.nextToken()
  }

  return block
}

/* parse Expressions */
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

// eg: true, false
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

  // 1.curToken is Prefix Token, jump it
  // !5;
  // -5;
  // ^..
  p.nextToken()

  // 2.recursive parsing
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

  // 1.curToken is Infix Token, jump it
  p.nextToken()

  // 2.recursive parsing
  expression.Right = p.parseExpression(precedence)

  return expression
}

// eg: (
func (p *Parser) parseGroupedExpression() ast.Expression {
  // 1.curToken is '(', jump it
  // (1 + 1)
  // ^......
  p.nextToken()

  // 2.Just raise the priority of the expressions inside the parens
  expression := p.parseExpression(LOWEST)

  // 3.peekToken may be ')'
  // (1 + 1)
  // ......^
  if !p.expectPeek(token.RPAREN) {
    return nil
  }
  // 4.curToken is ')'

  return expression
}

// eg: if (a > b) { a }
func (p *Parser) parseIfExpression() ast.Expression {
  expression := &ast.IfExpression{Token: p.curToken}

  // 1.curToken is 'if', peekToken may be '('
  // if (a > b) { a }
  // ...^............
  if !p.expectPeek(token.LPAREN) {
    return nil
  }

  // 2.curToken is '(', jump it
  p.nextToken()

  // 3.parseExpression
  expression.Condition = p.parseExpression(LOWEST)

  // 4.peekToken may be '('
  // if (a > b) { a }
  // .........^......
  if !p.expectPeek(token.RPAREN) {
    return nil
  }

  // 5.curToken is '(', peekToken may be '{'
  // if (a > b) { a }
  // ...........^...
  if !p.expectPeek(token.LBRACE) {
    return nil
  }

  // 6.curToken is '{'
  expression.Consequence = p.parseBlockStatement()

  // 7.peekToken may be 'else'
  // if (a > b) { a } else { b }
  // .................^^^^......
  if p.peekTokenIs(token.ELSE) {
    // 8.peekToken is 'else', jump it
    p.nextToken()

    // 9.curToken is 'else', peekToken may be '{'
    // if (a > b) { a } else { b }
    // ......................^...
    if !p.expectPeek(token.LBRACE) {
      return nil
    }

    // 10.curToken is '{'
    expression.Alternative = p.parseBlockStatement()
  }

  return expression
}

// eg: fn(a, b) { return a + b; }
func (p *Parser) parseFunctionLiteral() ast.Expression {
  literal := &ast.FunctionLiteral{Token: p.curToken}

  // 1.curToken is 'fn', peekToken may be '('
  // fn(a, b) { return a + b; }
  // ..^.......................
  if !p.expectPeek(token.LPAREN) {
    return nil
  }

  // 2.curToken is '(', parse Function Parameters
  literal.Parameters = p.parseFunctionParameters()

  // 3.curToken is ')', peekToken may be '{'
  // fn(a, b) { return a + b; }
  // .........^................
  if !p.expectPeek(token.LBRACE) {
    return nil
  }

  // 4.curToken is '{', parse BlockStatement
  literal.Body = p.parseBlockStatement()

  return literal
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
  identifiers := []*ast.Identifier{}

  // CASE 1: No Parameters, eg: fn()
  // 1.1 curToken is '(', peekToken may be ')'
  if p.peekTokenIs(token.RPAREN) {
    // 1.2 peekToken is ')', jump to it
    p.nextToken()
    // 1.3 curToken is ')'

    return identifiers
  }

  // CASE 2: Has Parameters, eg: fn(a, b, c)
  // 2.1 curToken is '(', jump it
  p.nextToken()

  // 2.2 first parameter
  // fn(a, b, c) {}
  // ...^..........
  identifier := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
  identifiers = append(identifiers, identifier)

  // 2.3 rest parameters
  // fn(a, b, c) {}
  // ......^^^^....
  for p.peekTokenIs(token.COMMA) {
    // peekToken is ',', jump to it
    p.nextToken()
    // curToken is ',', jump it
    p.nextToken()
    identifier := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
    identifiers = append(identifiers, identifier)
  }

  // 2.4 peekToken may be ')'
  // fn(a, b, c) {}
  // ..........^....
  if !p.expectPeek(token.RPAREN) {
    return nil
  }
  // 2.5 curToken is ')'

  return identifiers
}

// eg: add(1, 2 * 3, 4 + 5);
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
  expression := &ast.CallExpression{Token: p.curToken, Function: function}
  expression.Arguments = p.parseCallArguments()
  return expression
}

func (p *Parser) parseCallArguments() []ast.Expression {
  args := []ast.Expression{}

  // CASE 1: No Parameters, eg: add()
  // 1.1 curToken is '(', peekToken may be ')'
  if p.peekTokenIs(token.RPAREN) {
    // 1.2 peekToken is ')', jump to it
    p.nextToken()
    // 1.3 curToken is ')'

    return args
  }

  // CASE 2: Has Parameters, eg: add(a, b, c)
  // 2.1 curToken is '(', jump it
  p.nextToken()

  // 2.2 first arguments
  // add(a, b, c) {}
  // ....^..........
  args = append(args, p.parseExpression(LOWEST))

  // 2.3 rest parameters
  // add(a, b, c) {}
  // .......^^^^....
  for p.peekTokenIs(token.COMMA) {
    // peekToken is ',', jump to it
    p.nextToken()
    // curToken is ',', jump it
    p.nextToken()
    args = append(args, p.parseExpression(LOWEST))
  }

  // 2.4 peekToken may be ')'
  // add(a, b, c) {}
  // ...........^....
  if !p.expectPeek(token.RPAREN) {
    return nil
  }
  // 2.5 curToken is ')'

  return args
}

/* parse utils */
func (p *Parser) nextToken() {
  p.curToken = p.peekToken
  p.peekToken = p.l.NextToken()
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
