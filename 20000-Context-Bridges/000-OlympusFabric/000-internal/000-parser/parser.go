package parser

import (
	"fmt"

	"OlympusForge/20000-Context-Bridges/000-OlympusFabric/P0000-pkg/000-ir"
)

type Parser struct {
	l         *Lexer
	curToken  Token
	peekToken Token
	errors    []string
}

func NewParser(l *Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}
	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseGrammar() (*ir.Node, error) {
	root := ir.NewNode(ir.NodeRoot)

	for p.curToken.Type != TokenEOF {
		rule := p.parseRule()
		if rule != nil {
			root.AddChild(rule)
		}

		// Optional semicolon between rules
		if p.curToken.Type == TokenSemicolon {
			p.nextToken()
		}
	}

	if len(p.errors) > 0 {
		return nil, fmt.Errorf("parser errors: %v", p.errors)
	}

	return root, nil
}

func (p *Parser) parseRule() *ir.Node {
	if p.curToken.Type != TokenIdentifier {
		p.addError(fmt.Sprintf("expected identifier at line %d, got %s", p.curToken.Line, p.curToken.Literal))
		p.nextToken()
		return nil
	}

	ruleName := p.curToken.Literal
	p.nextToken()

	if p.curToken.Type != TokenEquals {
		p.addError(fmt.Sprintf("expected '=' after identifier %s", ruleName))
		return nil
	}
	p.nextToken()

	rhs := p.parseExpression()

	ruleNode := ir.NewNode(ir.NodeRule)
	ruleNode.SetAttribute("name", ruleName)
	ruleNode.AddChild(rhs)

	return ruleNode
}

func (p *Parser) parseExpression() *ir.Node {
	// Root of expression is usually a Choice (ordered or unordered)
	choices := make([]*ir.Node, 0)
	isOrdered := false

	choices = append(choices, p.parseSequence())

	for p.curToken.Type == TokenPipe || p.curToken.Type == TokenSlash {
		if p.curToken.Type == TokenSlash {
			isOrdered = true
		}
		p.nextToken()
		choices = append(choices, p.parseSequence())
	}

	if len(choices) == 1 {
		return choices[0]
	}

	choiceNode := ir.NewNode(ir.NodeChoice)
	if isOrdered {
		choiceNode.SetAttribute("ordered", true)
	}
	for _, c := range choices {
		choiceNode.AddChild(c)
	}
	return choiceNode
}

func (p *Parser) parseSequence() *ir.Node {
	items := make([]*ir.Node, 0)

	for {
		item := p.parseTerm()
		if item == nil {
			break
		}
		items = append(items, item)

		// Optional comma in some EBNF variants
		if p.curToken.Type == TokenComma {
			p.nextToken()
		}

		// Stop if we hit a boundary
		if p.curToken.Type == TokenPipe || p.curToken.Type == TokenSlash ||
			p.curToken.Type == TokenSemicolon || p.curToken.Type == TokenEOF ||
			p.curToken.Type == TokenRParen || p.curToken.Type == TokenRBracket ||
			p.curToken.Type == TokenRBrace {
			break
		}
	}

	if len(items) == 1 {
		return items[0]
	}

	seqNode := ir.NewNode(ir.NodeSequence)
	for _, item := range items {
		seqNode.AddChild(item)
	}
	return seqNode
}

func (p *Parser) parseTerm() *ir.Node {
	var node *ir.Node

	switch p.curToken.Type {
	case TokenIdentifier:
		node = ir.NewNode(ir.NodeReference)
		node.SetAttribute("name", p.curToken.Literal)
		p.nextToken()
	case TokenLiteral:
		node = ir.NewNode(ir.NodeLiteral)
		node.SetAttribute("value", p.curToken.Literal)
		p.nextToken()
		if p.curToken.Type == TokenRange {
			p.nextToken()
			if p.curToken.Type == TokenLiteral {
				node.SetAttribute("range_to", p.curToken.Literal)
				p.nextToken()
			}
		}
	case TokenLParen:
		p.nextToken()
		node = p.parseExpression()
		if p.curToken.Type != TokenRParen {
			p.addError("missing closing parenthesis")
		} else {
			p.nextToken()
		}
	case TokenLBracket:
		p.nextToken()
		inner := p.parseExpression()
		node = ir.NewNode(ir.NodeOptional)
		node.AddChild(inner)
		if p.curToken.Type != TokenRBracket {
			p.addError("missing closing bracket")
		} else {
			p.nextToken()
		}
	case TokenLBrace:
		p.nextToken()
		inner := p.parseExpression()
		node = ir.NewNode(ir.NodeRepeat)
		node.AddChild(inner)
		if p.curToken.Type != TokenRBrace {
			p.addError("missing closing brace")
		} else {
			p.nextToken()
		}
	default:
		return nil
	}

	// Handle postfix operators (?, *, +)
	for {
		if p.curToken.Type == TokenQuestion {
			opt := ir.NewNode(ir.NodeOptional)
			opt.AddChild(node)
			node = opt
			p.nextToken()
		} else if p.curToken.Type == TokenStar {
			rep := ir.NewNode(ir.NodeRepeat)
			rep.AddChild(node)
			node = rep
			p.nextToken()
		} else if p.curToken.Type == TokenPlus {
			rep := ir.NewNode(ir.NodeRepeat)
			rep.SetAttribute("at_least_one", true)
			rep.AddChild(node)
			node = rep
			p.nextToken()
		} else {
			break
		}
	}

	return node
}

func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, msg)
}
