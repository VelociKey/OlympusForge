package parser

import (
	"fmt"

	"unicode"
)

type TokenType int

const (
	TokenError TokenType = iota
	TokenEOF
	TokenIdentifier
	TokenLiteral
	TokenEquals    // =
	TokenPipe      // |
	TokenSlash     // / (Ordered choice in jeBNF)
	TokenSemicolon // ;
	TokenLParen    // (
	TokenRParen    // )
	TokenLBracket  // [
	TokenRBracket  // ]
	TokenLBrace    // {
	TokenRBrace    // }
	TokenStar      // *
	TokenPlus      // +
	TokenQuestion  // ?
	TokenComma     // ,
	TokenRange     // ..
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Pos     int
}

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           rune
	line         int
	posInLine    int
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input, line: 1}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = rune(l.input[l.readPosition])
	}
	l.position = l.readPosition
	l.readPosition++
	l.posInLine++
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	var tok Token
	tok.Line = l.line
	tok.Pos = l.posInLine

	switch l.ch {
	case '=':
		tok.Type = TokenEquals
		tok.Literal = string(l.ch)
	case '|':
		tok.Type = TokenPipe
		tok.Literal = string(l.ch)
	case '/':
		if l.peekChar() == '*' {
			l.skipCStyleComment()
			return l.NextToken()
		}
		tok.Type = TokenSlash
		tok.Literal = string(l.ch)
	case ';':
		tok.Type = TokenSemicolon
		tok.Literal = string(l.ch)
	case '(':
		if l.peekChar() == '*' {
			l.skipComment()
			return l.NextToken()
		}
		tok.Type = TokenLParen
		tok.Literal = string(l.ch)
	case ')':
		tok.Type = TokenRParen
		tok.Literal = string(l.ch)
	case '[':
		tok.Type = TokenLBracket
		tok.Literal = string(l.ch)
	case ']':
		tok.Type = TokenRBracket
		tok.Literal = string(l.ch)
	case '{':
		tok.Type = TokenLBrace
		tok.Literal = string(l.ch)
	case '}':
		tok.Type = TokenRBrace
		tok.Literal = string(l.ch)
	case '*':
		tok.Type = TokenStar
		tok.Literal = string(l.ch)
	case '+':
		tok.Type = TokenPlus
		tok.Literal = string(l.ch)
	case '?':
		tok.Type = TokenQuestion
		tok.Literal = string(l.ch)
	case ',':
		tok.Type = TokenComma
		tok.Literal = string(l.ch)
	case '.':
		if l.peekChar() == '.' {
			l.readChar()
			tok.Type = TokenRange

			tok.Literal = ".."
		} else {
			tok.Type = TokenError

			tok.Literal = "."
		}

	case '"', '\'':
		tok.Type = TokenLiteral
		tok.Literal = l.readLiteral(l.ch)
		return tok
	case 0:
		tok.Type = TokenEOF
		tok.Literal = ""
	default:
		if isLetter(l.ch) || unicode.IsDigit(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = TokenIdentifier
			return tok
		} else {
			tok.Type = TokenError
			tok.Literal = string(l.ch)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.ch) {
		if l.ch == '\n' {
			l.line++
			l.posInLine = 0
		}
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	// Skip (*
	l.readChar()
	l.readChar()
	for l.ch != 0 {
		if l.ch == '*' && l.peekChar() == ')' {
			l.readChar()
			l.readChar()
			return
		}
		if l.ch == '\n' {
			l.line++
			l.posInLine = 0
		}
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || unicode.IsDigit(l.ch) || l.ch == '_' || l.ch == '-' || l.ch == '.' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readLiteral(quote rune) string {
	l.readChar() // skip starting quote
	position := l.position
	for l.ch != quote && l.ch != 0 {
		if l.ch == '\\' {
			l.readChar() // skip escape
		}
		l.readChar()
	}
	literal := l.input[position:l.position]
	l.readChar() // skip ending quote
	return literal
}

func (l *Lexer) peekChar() rune {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return rune(l.input[l.readPosition])
}

func isLetter(ch rune) bool {
	return unicode.IsLetter(ch)
}

func (t Token) String() string {

	return fmt.Sprintf("Token(%d, %q, line %d, pos %d)", t.Type, t.Literal, t.Line, t.Pos)
}

func (l *Lexer) skipCStyleComment() {
	// Skip /*
	l.readChar()
	l.readChar()
	for l.ch != 0 {
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar()
			l.readChar()
			return
		}
		if l.ch == '\n' {
			l.line++
			l.posInLine = 0
		}
		l.readChar()
	}
}
