package lexer

import (
  "JFFMonkeyLang/src/token"
)

type Lexer struct {
  input string
  // current position in input (points to current char)
  position int
  // current reading position in input (after current char)
  // readPosition = position + 1
  readPosition int
  // current char under examination
  ch byte
}

func New(input string) *Lexer {
  l := &Lexer{input: input}
  l.readChar()
  return l
}

func (l *Lexer) NextToken() token.Token {
  var tok token.Token

  l.skipWhitespace()

  switch l.ch {
  case '=':
    // '==' token
    if l.peekChar() == '=' {
      l.readChar()

      tok.Literal = "=="
      tok.Type = token.EQ
    } else {
      // '=' token
      tok = newToken(token.ASSIGN, l.ch)
    }
  case '!':
    // '!=' token
    if l.peekChar() == '=' {
      l.readChar()

      tok.Literal = "!="
      tok.Type = token.NOT_EQ
    } else {
      // '!' token
      tok = newToken(token.BANG, l.ch)
    }
  case '+':
    tok = newToken(token.PLUS, l.ch)
  case '-':
    tok = newToken(token.MINUS, l.ch)
  case '*':
    tok = newToken(token.ASTERISK, l.ch)
  case '/':
    tok = newToken(token.SLASH, l.ch)
  case '<':
    tok = newToken(token.LT, l.ch)
  case '>':
    tok = newToken(token.GT, l.ch)
  case '(':
    tok = newToken(token.LPAREN, l.ch)
  case ')':
    tok = newToken(token.RPAREN, l.ch)
  case '{':
    tok = newToken(token.LBRACE, l.ch)
  case '}':
    tok = newToken(token.RBRACE, l.ch)
  case ',':
    tok = newToken(token.COMMA, l.ch)
  case ';':
    tok = newToken(token.SEMICOLON, l.ch)
  case 0:
    tok.Literal = ""
    tok.Type = token.EOF
  default:
    if isLetter(l.ch) {
      tok.Literal = l.readIdentifier()
      tok.Type = token.LookupIdent(tok.Literal)
      return tok
    }

    if isDigit(l.ch) {
      tok.Literal = l.readNumber()
      tok.Type = token.INT
      return tok
    }

    // unkown token or illegal token
    tok = newToken(token.ILLEGAL, l.ch)
  }

  l.readChar()
  return tok
}

func (l *Lexer) readChar() {
  if (l.readPosition) >= len(l.input) {
    l.ch = 0 // 0 is NULL ASCII code
  } else {
    l.ch = l.input[l.readPosition]
  }
  l.position = l.readPosition
  l.readPosition += 1
}

func (l *Lexer) readIdentifier() string {
  position := l.position
  for isLetter(l.ch) {
    l.readChar()
  }

  return l.input[position:l.position]
}

func (l *Lexer) peekChar() byte {
  // check edge cases
  if l.readPosition >= len(l.input) {
    return 0
  }

  return l.input[l.readPosition]
}

func (l *Lexer) readNumber() string {
  position := l.position
  for isDigit(l.ch) {
    l.readChar()
  }

  return l.input[position:l.position]
}

func (l *Lexer) skipWhitespace() {
  for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
    l.readChar()
  }
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
  return token.Token{Type: tokenType, Literal: string(ch)}
}

func isLetter(ch byte) bool {
  // a-zA-Z_
  return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_'
}

func isDigit(ch byte) bool {
  // 0-9
  return '0' <= ch && ch <= '9'
}
