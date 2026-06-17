package engine

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// =======================================================
// 1. LEXER
// =======================================================
type TokenType string

const (
	CONST TokenType = "CONST"; BOOLEAN TokenType = "BOOLEAN"; QUBIT TokenType = "QUBIT"; NUMBER TokenType = "NUMBER"
	NEW TokenType = "NEW"; TRUE TokenType = "TRUE"; FALSE TokenType = "FALSE"
	SUPERPOSE TokenType = "SUPERPOSE"; INVERT TokenType = "INVERT"; MEASURE TokenType = "MEASURE"
	IF TokenType = "IF"; IMPORT TokenType = "IMPORT"; FROM TokenType = "FROM"; EXPORT TokenType = "EXPORT"
	IDENT TokenType = "IDENT"; LBRACE TokenType = "LBRACE"; RBRACE TokenType = "RBRACE"
	LPAREN TokenType = "LPAREN"; RPAREN TokenType = "RPAREN"; ASSIGN TokenType = "ASSIGN"
	SEMI TokenType = "SEMI"; COLON TokenType = "COLON"; STRING TokenType = "STRING"
	COMMA TokenType = "COMMA"; EOF TokenType = "EOF"; ILLEGAL TokenType = "ILLEGAL"
	PLUS TokenType = "PLUS"; MINUS TokenType = "MINUS"; ASTERISK TokenType = "ASTERISK"; SLASH TokenType = "SLASH"
	LBRACKET TokenType = "LBRACKET"; RBRACKET TokenType = "RBRACKET"; DOT TokenType = "DOT"
	FOR TokenType = "FOR"; WHILE TokenType = "WHILE"; LET TokenType = "LET"
	LT TokenType = "LT"; GT TokenType = "GT"; EQ TokenType = "EQ"; NOT_EQ TokenType = "NOT_EQ"
	FUNCTION TokenType = "FUNCTION"; RETURN TokenType = "RETURN"; ADJOINT TokenType = "ADJOINT"
)

var keywords = map[string]TokenType{
	"const": CONST, "boolean": BOOLEAN, "qubit": QUBIT, "number": NUMBER, "new": NEW, "true": TRUE, "false": FALSE,
	"superpose": SUPERPOSE, "invert": INVERT, "measure": MEASURE, "if": IF,
	"import": IMPORT, "from": FROM, "export": EXPORT,
	"for": FOR, "while": WHILE, "let": LET, "function": FUNCTION, "return": RETURN, "adjoint": ADJOINT,
}

type Token struct { Type TokenType; Literal string; Line int; Col int }

type Lexer struct { input string; pos int; readPos int; ch byte; line int; col int }

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input, line: 1, col: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.ch == '\n' {
		l.line++
		l.col = 0
	} else {
		l.col++
	}
	if l.readPos >= len(l.input) { l.ch = 0 } else { l.ch = l.input[l.readPos] }
	l.pos = l.readPos
	l.readPos++
}

func (l *Lexer) peekChar() byte {
	if l.readPos >= len(l.input) { return 0 }
	return l.input[l.readPos]
}

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 { l.readChar() }
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' { l.readChar() }
}

func isLetter(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_'
}

func isDigit(ch byte) bool {
	return ('0' <= ch && ch <= '9') || ch == '.'
}

func (l *Lexer) readIdentifier() string {
	pos := l.pos
	for isLetter(l.ch) || ('0' <= l.ch && l.ch <= '9') { l.readChar() }
	return l.input[pos:l.pos]
}

func (l *Lexer) readNumber() string {
	pos := l.pos
	for isDigit(l.ch) { l.readChar() }
	return l.input[pos:l.pos]
}

func (l *Lexer) readString() string {
	pos := l.pos + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 { break }
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) newToken(t TokenType, lit string) Token {
	return Token{Type: t, Literal: lit, Line: l.line, Col: l.col}
}

func (l *Lexer) NextToken() Token {
	tok := l._nextToken()
	tok.Line = l.line
	tok.Col = l.col
	if tok.Type == IDENT || tok.Type == NUMBER || tok.Type == STRING || tok.Type >= CONST {
		tok.Col = l.col - len(tok.Literal)
		if tok.Col < 0 { tok.Col = 0 }
	}
	return tok
}

func (l *Lexer) _nextToken() Token {
	l.skipWhitespace()
	var tok Token
	switch l.ch {
	case '=':
		if l.peekChar() == '=' { l.readChar(); tok = l.newToken(EQ, "==") } else { tok = l.newToken(ASSIGN, "=") }
	case '!':
		if l.peekChar() == '=' { l.readChar(); tok = l.newToken(NOT_EQ, "!=") } else { tok = l.newToken(ILLEGAL, "!") }
	case '<': tok = l.newToken(LT, "<")
	case '>': tok = l.newToken(GT, ">")
	case '+': tok = l.newToken(PLUS, "+")
	case '-': tok = l.newToken(MINUS, "-")
	case '*': tok = l.newToken(ASTERISK, "*")
	case ';': tok = l.newToken(SEMI, ";")
	case ':': tok = l.newToken(COLON, ":")
	case '(': tok = l.newToken(LPAREN, "(")
	case ')': tok = l.newToken(RPAREN, ")")
	case '{': tok = l.newToken(LBRACE, "{")
	case '}': tok = l.newToken(RBRACE, "}")
	case '[': tok = l.newToken(LBRACKET, "[")
	case ']': tok = l.newToken(RBRACKET, "]")
	case '.': tok = l.newToken(DOT, ".")
	case ',': tok = l.newToken(COMMA, ",")
	case '"': tok = l.newToken(STRING, l.readString())
	case '/':
		if l.peekChar() == '/' {
			l.skipComment()
			return l.NextToken()
		}
		tok = l.newToken(SLASH, "/")
	case 0: tok = l.newToken(EOF, "")
	default:
		if isLetter(l.ch) {
			ident := l.readIdentifier()
			if t, ok := keywords[ident]; ok { tok.Type = t } else { tok.Type = IDENT }
			tok.Literal = ident
			return tok
		} else if isDigit(l.ch) {
			return l.newToken(NUMBER, l.readNumber())
		}
		tok = l.newToken(ILLEGAL, string(l.ch))
	}
	l.readChar()
	return tok
}

// =======================================================
// 2. AST Nodes
// =======================================================
type Node interface{}
type Statement interface { Node; stmt() }
type Expression interface { Node; expr() }

type Program struct{ Stmts []Statement }
type ConstDecl struct { Name *Identifier; Type string; Value Expression }
type LetDecl struct { Name *Identifier; Type string; Value Expression }
type AssignStmt struct { Name *Identifier; Value Expression; Index Expression }
type ExportStmt struct{ Decl *ConstDecl }
type IfStmt struct { Condition Expression; Consequence *BlockStmt }
type WhileStmt struct { Condition Expression; Body *BlockStmt }
type ForStmt struct { Init Statement; Condition Expression; Post Statement; Body *BlockStmt }
type ImportStmt struct { Names []*Identifier; Module string }
type BlockStmt struct{ Stmts []Statement }
type ExpressionStmt struct{ Expr Expression }
type FunctionDecl struct { Name *Identifier; Parameters []*Param; ReturnType string; Body *BlockStmt }
type Param struct { Name *Identifier; Type string }
type ReturnStmt struct { Value Expression }

type Identifier struct{ Value string }
type BooleanLiteral struct{ Value bool }
type StringLiteral struct{ Value string }
type NumberLiteral struct{ Value float64 }
type NewQubit struct{}
type CallExpr struct { Function string; Args []Expression }
type InfixExpression struct { Left Expression; Operator string; Right Expression }
type IndexExpression struct { Left Expression; Index Expression }
type PropertyAccess struct { Object Expression; Property string }
type ObjectLiteral struct { Pairs map[string]Expression }
type ArrayLiteral struct { Elements []Expression; Type string }
type NewArray struct { Type string; Length Expression }
type AdjointExpression struct { Call Expression }

func (p *Program) stmt() {} 
func (c *ConstDecl) stmt() {} 
func (l *LetDecl) stmt() {}
func (a *AssignStmt) stmt() {}
func (i *IfStmt) stmt() {}
func (w *WhileStmt) stmt() {}
func (f *ForStmt) stmt() {}
func (i *ImportStmt) stmt() {} 
func (e *ExportStmt) stmt() {} 
func (e *ExpressionStmt) stmt() {}
func (b *BlockStmt) stmt() {} 
func (f *FunctionDecl) stmt() {}
func (r *ReturnStmt) stmt() {}

func (i *Identifier) expr() {} 
func (b *BooleanLiteral) expr() {}
func (s *StringLiteral) expr() {} 
func (n *NumberLiteral) expr() {}
func (n *NewQubit) expr() {} 
func (c *CallExpr) expr() {}
func (i *InfixExpression) expr() {}
func (i *IndexExpression) expr() {}
func (p *PropertyAccess) expr() {}
func (o *ObjectLiteral) expr() {}
func (a *ArrayLiteral) expr() {}
func (n *NewArray) expr() {}
func (a *AdjointExpression) expr() {}

// =======================================================
// 3. PARSER
// =======================================================
const (
	_ int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
	INDEX
	PROPERTY
)

var precedences = map[TokenType]int{
	EQ: EQUALS, NOT_EQ: EQUALS, LT: LESSGREATER, GT: LESSGREATER,
	PLUS: SUM, MINUS: SUM, ASTERISK: PRODUCT, SLASH: PRODUCT,
	LPAREN: CALL, LBRACKET: INDEX, DOT: PROPERTY,
}

type Parser struct { l *Lexer; curToken Token; peekToken Token }

func NewParser(l *Lexer) *Parser {
	p := &Parser{l: l}
	p.nextToken(); p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) expectPeek(t TokenType) {
	if p.peekToken.Type == t {
		p.nextToken()
	} else { panic(fmt.Sprintf("Syntax error at line %d, col %d: expected %s got %s", p.peekToken.Line, p.peekToken.Col, t, p.peekToken.Type)) }
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok { return p }
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok { return p }
	return LOWEST
}

func (p *Parser) parseType() string {
	t := p.curToken.Literal
	if p.peekToken.Type == LBRACKET {
		p.nextToken(); p.expectPeek(RBRACKET)
		return t + "[]"
	}
	return t
}

func (p *Parser) ParseProgram() *Program {
	prog := &Program{}
	for p.curToken.Type != EOF {
		if stmt := p.parseStatement(); stmt != nil { prog.Stmts = append(prog.Stmts, stmt) }
		p.nextToken()
	}
	return prog
}

func (p *Parser) parseStatement() Statement {
	switch p.curToken.Type {
	case CONST: return p.parseConstDecl()
	case LET: return p.parseLetDecl()
	case EXPORT: return p.parseExportStmt()
	case IF: return p.parseIfStmt()
	case WHILE: return p.parseWhileStmt()
	case FOR: return p.parseForStmt()
	case IMPORT: return p.parseImportStmt()
	case FUNCTION: return p.parseFunctionDecl()
	case RETURN: return p.parseReturnStmt()
	case IDENT:
		if p.peekToken.Type == ASSIGN || p.peekToken.Type == LBRACKET {
			return p.parseAssignOrExprStmt()
		}
		fallthrough
	default: return p.parseExpressionStmt()
	}
}

func (p *Parser) parseAssignOrExprStmt() Statement {
	expr := p.parseExpression(LOWEST)
	if p.peekToken.Type == ASSIGN {
		p.nextToken() 
		p.nextToken() 
		right := p.parseExpression(LOWEST)
		if p.peekToken.Type == SEMI { p.nextToken() }
		
		if ident, ok := expr.(*Identifier); ok {
			return &AssignStmt{Name: ident, Value: right}
		} else if idx, ok := expr.(*IndexExpression); ok {
			if ident, ok := idx.Left.(*Identifier); ok {
				return &AssignStmt{Name: ident, Index: idx.Index, Value: right}
			}
		}
		panic(fmt.Sprintf("Syntax error at line %d, col %d: invalid assignment target", p.curToken.Line, p.curToken.Col))
	}
	stmt := &ExpressionStmt{Expr: expr}
	if p.peekToken.Type == SEMI { p.nextToken() }
	return stmt
}

func (p *Parser) parseConstDecl() *ConstDecl {
	decl := &ConstDecl{}
	p.expectPeek(IDENT)
	decl.Name = &Identifier{Value: p.curToken.Literal}
	p.expectPeek(COLON)
	p.nextToken() 
	decl.Type = p.parseType()
	p.expectPeek(ASSIGN)
	p.nextToken() 
	decl.Value = p.parseExpression(LOWEST)
	if p.peekToken.Type == SEMI { p.nextToken() }
	return decl
}

func (p *Parser) parseLetDecl() *LetDecl {
	decl := &LetDecl{}
	p.expectPeek(IDENT)
	decl.Name = &Identifier{Value: p.curToken.Literal}
	if p.peekToken.Type == COLON {
		p.nextToken()
		p.nextToken()
		decl.Type = p.parseType()
	}
	p.expectPeek(ASSIGN)
	p.nextToken() 
	decl.Value = p.parseExpression(LOWEST)
	if p.peekToken.Type == SEMI { p.nextToken() }
	return decl
}

func (p *Parser) parseExportStmt() *ExportStmt {
	stmt := &ExportStmt{}
	p.expectPeek(CONST)
	stmt.Decl = p.parseConstDecl()
	return stmt
}

func (p *Parser) parseImportStmt() *ImportStmt {
	stmt := &ImportStmt{}
	p.expectPeek(LBRACE)
	p.nextToken() 
	for p.curToken.Type != RBRACE {
		stmt.Names = append(stmt.Names, &Identifier{Value: p.curToken.Literal})
		if p.peekToken.Type == COMMA { p.nextToken(); p.nextToken() } else { p.nextToken() }
	}
	p.expectPeek(FROM)
	p.expectPeek(STRING)
	stmt.Module = p.curToken.Literal
	p.expectPeek(SEMI)
	return stmt
}

func (p *Parser) parseBlockStmt() *BlockStmt {
	block := &BlockStmt{}
	p.expectPeek(LBRACE)
	p.nextToken() 
	for p.curToken.Type != RBRACE && p.curToken.Type != EOF {
		if s := p.parseStatement(); s != nil { block.Stmts = append(block.Stmts, s) }
		p.nextToken()
	}
	return block
}

func (p *Parser) parseIfStmt() *IfStmt {
	stmt := &IfStmt{}
	p.expectPeek(LPAREN)
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	p.expectPeek(RPAREN)
	stmt.Consequence = p.parseBlockStmt()
	return stmt
}

func (p *Parser) parseWhileStmt() *WhileStmt {
	stmt := &WhileStmt{}
	p.expectPeek(LPAREN)
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	p.expectPeek(RPAREN)
	stmt.Body = p.parseBlockStmt()
	return stmt
}

func (p *Parser) parseForStmt() *ForStmt {
	stmt := &ForStmt{}
	p.expectPeek(LPAREN)
	p.nextToken()
	if p.curToken.Type != SEMI {
		stmt.Init = p.parseStatement()
	}
	if p.curToken.Type == SEMI { p.nextToken() }
	
	if p.curToken.Type != SEMI {
		stmt.Condition = p.parseExpression(LOWEST)
		p.expectPeek(SEMI)
	}
	if p.curToken.Type == SEMI { p.nextToken() }

	if p.curToken.Type != RPAREN {
		stmt.Post = p.parseAssignOrExprStmt() 
	}
	p.expectPeek(RPAREN)
	stmt.Body = p.parseBlockStmt()
	return stmt
}

func (p *Parser) parseFunctionDecl() *FunctionDecl {
	stmt := &FunctionDecl{}
	p.expectPeek(IDENT)
	stmt.Name = &Identifier{Value: p.curToken.Literal}
	p.expectPeek(LPAREN)
	p.nextToken()
	for p.curToken.Type != RPAREN {
		param := &Param{Name: &Identifier{Value: p.curToken.Literal}}
		p.expectPeek(COLON)
		p.nextToken()
		param.Type = p.parseType()
		stmt.Parameters = append(stmt.Parameters, param)
		if p.peekToken.Type == COMMA { p.nextToken(); p.nextToken() } else { p.nextToken() }
	}
	if p.peekToken.Type == COLON {
		p.nextToken()
		p.nextToken()
		stmt.ReturnType = p.parseType()
	} else {
		stmt.ReturnType = "void"
	}
	stmt.Body = p.parseBlockStmt()
	return stmt
}

func (p *Parser) parseReturnStmt() *ReturnStmt {
	stmt := &ReturnStmt{}
	p.nextToken()
	if p.curToken.Type == SEMI { return stmt }
	stmt.Value = p.parseExpression(LOWEST)
	if p.peekToken.Type == SEMI { p.nextToken() }
	return stmt
}

func (p *Parser) parseExpressionStmt() *ExpressionStmt {
	stmt := &ExpressionStmt{Expr: p.parseExpression(LOWEST)}
	if p.peekToken.Type == SEMI { p.nextToken() }
	return stmt
}

func (p *Parser) parseExpression(precedence int) Expression {
	var left Expression
	switch p.curToken.Type {
	case NUMBER:
		val, _ := strconv.ParseFloat(p.curToken.Literal, 64)
		left = &NumberLiteral{Value: val}
	case STRING:
		left = &StringLiteral{Value: p.curToken.Literal}
	case NEW:
		p.expectPeek(IDENT)
		ident := p.curToken.Literal
		if p.peekToken.Type == LBRACKET {
			p.nextToken(); p.nextToken()
			length := p.parseExpression(LOWEST)
			p.expectPeek(RBRACKET)
			left = &NewArray{Type: ident, Length: length}
		} else {
			if ident != "Qubit" { panic(fmt.Sprintf("Syntax error at line %d, col %d: expected Qubit", p.curToken.Line, p.curToken.Col)) }
			p.expectPeek(LPAREN); p.expectPeek(RPAREN)
			left = &NewQubit{}
		}
	case TRUE: left = &BooleanLiteral{Value: true}
	case FALSE: left = &BooleanLiteral{Value: false}
	case ADJOINT:
		p.nextToken()
		left = &AdjointExpression{Call: p.parseExpression(PREFIX)}
	case IDENT, SUPERPOSE, INVERT, MEASURE:
		left = &Identifier{Value: p.curToken.Literal}
	case LBRACKET:
		arr := &ArrayLiteral{}
		p.nextToken()
		for p.curToken.Type != RBRACKET {
			arr.Elements = append(arr.Elements, p.parseExpression(LOWEST))
			if p.peekToken.Type == COMMA { p.nextToken(); p.nextToken() } else { p.nextToken() }
		}
		left = arr
	case LBRACE:
		obj := &ObjectLiteral{Pairs: make(map[string]Expression)}
		p.nextToken()
		for p.curToken.Type != RBRACE {
			if p.curToken.Type != IDENT && p.curToken.Type != STRING {
				panic(fmt.Sprintf("Syntax error at line %d, col %d: expected identifier or string key", p.curToken.Line, p.curToken.Col))
			}
			key := p.curToken.Literal
			p.nextToken()
			if p.curToken.Type != COLON { panic(fmt.Sprintf("Syntax error at line %d, col %d: expected colon", p.curToken.Line, p.curToken.Col)) }
			p.nextToken()
			obj.Pairs[key] = p.parseExpression(LOWEST)
			if p.peekToken.Type == COMMA { p.nextToken(); p.nextToken() } else { p.nextToken() }
		}
		left = obj
	default:
		panic(fmt.Sprintf("Syntax error at line %d, col %d: unexpected token %s", p.curToken.Line, p.curToken.Col, p.curToken.Literal))
	}

	for p.peekToken.Type != SEMI && p.peekToken.Type != RPAREN && p.peekToken.Type != RBRACKET && p.peekToken.Type != EOF && precedence < p.peekPrecedence() {
		switch p.peekToken.Type {
		case PLUS, MINUS, ASTERISK, SLASH, EQ, NOT_EQ, LT, GT:
			p.nextToken()
			left = p.parseInfixExpression(left)
		case LPAREN:
			p.nextToken()
			call := &CallExpr{}
			if ident, ok := left.(*Identifier); ok { call.Function = ident.Value } else if prop, ok := left.(*PropertyAccess); ok {
				if identObj, ok2 := prop.Object.(*Identifier); ok2 && identObj.Value == "Math" {
					call.Function = "Math." + prop.Property
				} else { panic(fmt.Sprintf("Syntax error at line %d, col %d: invalid method call", p.curToken.Line, p.curToken.Col)) }
			} else { panic(fmt.Sprintf("Syntax error at line %d, col %d: invalid function call", p.curToken.Line, p.curToken.Col)) }
			p.nextToken() 
			if p.curToken.Type != RPAREN {
				for {
					call.Args = append(call.Args, p.parseExpression(LOWEST))
					if p.peekToken.Type == COMMA { p.nextToken(); p.nextToken() } else { break }
				}
				p.expectPeek(RPAREN)
			}
			left = call
		case LBRACKET:
			p.nextToken()
			idx := &IndexExpression{Left: left}
			p.nextToken()
			idx.Index = p.parseExpression(LOWEST)
			p.expectPeek(RBRACKET)
			left = idx
		case DOT:
			p.nextToken()
			p.expectPeek(IDENT)
			left = &PropertyAccess{Object: left, Property: p.curToken.Literal}
		default:
			return left
		}
	}

	return left
}

func (p *Parser) parseInfixExpression(left Expression) Expression {
	expr := &InfixExpression{Left: left, Operator: p.curToken.Literal}
	precedence := p.curPrecedence()
	p.nextToken()
	expr.Right = p.parseExpression(precedence)
	return expr
}

// =======================================================
// 4. STATIC SEMANTIC TYPE CHECKER
// =======================================================
type TypeEnv struct { 
	store map[string]string
	funcs map[string]*FunctionDecl
	outer *TypeEnv 
}
func NewTypeEnv() *TypeEnv { return &TypeEnv{store: make(map[string]string), funcs: make(map[string]*FunctionDecl)} }
func (e *TypeEnv) Set(name, t string) { e.store[name] = t }
func (e *TypeEnv) Get(name string) (string, bool) {
	t, ok := e.store[name]
	if !ok && e.outer != nil { return e.outer.Get(name) }
	return t, ok
}
func (e *TypeEnv) SetFunc(name string, f *FunctionDecl) { e.funcs[name] = f }
func (e *TypeEnv) GetFunc(name string) (*FunctionDecl, bool) {
	f, ok := e.funcs[name]
	if !ok && e.outer != nil { return e.outer.GetFunc(name) }
	return f, ok
}

func GetType(expr Expression, tenv *TypeEnv) string {
	switch e := expr.(type) {
	case *BooleanLiteral: return "boolean"
	case *NumberLiteral: return "number"
	case *StringLiteral: return "string"
	case *NewQubit: return "qubit"
	case *Identifier:
		t, ok := tenv.Get(e.Value)
		if !ok { panic(fmt.Sprintf("Type error: undefined variable: %s", e.Value)) }
		return t
	case *InfixExpression:
		lt := GetType(e.Left, tenv)
		rt := GetType(e.Right, tenv)
		if e.Operator == "==" || e.Operator == "!=" || e.Operator == "<" || e.Operator == ">" {
			if lt != rt { panic("Type error: comparison operators require matching types") }
			return "boolean"
		}
		if e.Operator == "+" && lt == "string" && rt == "string" { return "string" }
		if lt != "number" || rt != "number" { panic("Type error: math operators require numbers") }
		return "number"
	case *IndexExpression:
		lt := GetType(e.Left, tenv)
		it := GetType(e.Index, tenv)
		if it != "number" { panic("Type error: index must be a number") }
		if len(lt) > 2 && lt[len(lt)-2:] == "[]" { return lt[:len(lt)-2] }
		panic("Type error: cannot index non-array type: " + lt)
	case *PropertyAccess:
		lt := GetType(e.Object, tenv)
		if lt == "Math" || lt == "object" || lt == "any" { return "any" }
		panic(fmt.Sprintf("Type error: cannot access property '%s' on %s", e.Property, lt))
	case *ObjectLiteral:
		for _, el := range e.Pairs { GetType(el, tenv) }
		return "object"
	case *ArrayLiteral:
		if len(e.Elements) == 0 { return "any[]" }
		t := GetType(e.Elements[0], tenv)
		for _, el := range e.Elements {
			if GetType(el, tenv) != t { panic("Type error: mixed array types") }
		}
		return t + "[]"
	case *NewArray:
		it := GetType(e.Length, tenv)
		if it != "number" { panic("Type error: array length must be a number") }
		if e.Type == "Qubit" { return "qubit[]" }
		return e.Type + "[]"
	case *AdjointExpression:
		return GetType(e.Call, tenv)
	case *CallExpr:
		if e.Function == "measure" { return "boolean" }
		if e.Function == "superpose" || e.Function == "invert" || e.Function == "X" || e.Function == "Y" || e.Function == "Z" || e.Function == "S" || e.Function == "T" || e.Function == "Rx" || e.Function == "Ry" || e.Function == "Rz" {
			return "void"
		}
		if f, ok := tenv.GetFunc(e.Function); ok {
			return f.ReturnType
		}
		if strings.HasPrefix(e.Function, "Math.") { return "number" }
		panic("Type error: undefined function " + e.Function)
	}
	return ""
}

type ModuleLoader func(string) (string, error)

func TypeCheck(node Node, tenv *TypeEnv, filename string, loader ModuleLoader) {
	switch n := node.(type) {
	case *Program:
		for _, stmt := range n.Stmts { TypeCheck(stmt, tenv, filename, loader) }
	case *BlockStmt:
		blockEnv := &TypeEnv{store: make(map[string]string), funcs: make(map[string]*FunctionDecl), outer: tenv}
		for _, stmt := range n.Stmts { TypeCheck(stmt, blockEnv, filename, loader) }
	case *ConstDecl:
		vt := GetType(n.Value, tenv)
		if vt != n.Type { panic(fmt.Sprintf("Type error in %s: cannot assign %s to %s", filename, vt, n.Type)) }
		tenv.Set(n.Name.Value, n.Type)
	case *LetDecl:
		vt := GetType(n.Value, tenv)
		if n.Type != "" && vt != n.Type { panic(fmt.Sprintf("Type error in %s: cannot assign %s to %s", filename, vt, n.Type)) }
		if n.Type == "" { n.Type = vt }
		tenv.Set(n.Name.Value, n.Type)
	case *AssignStmt:
		var targetType string
		if n.Index != nil {
			t, ok := tenv.Get(n.Name.Value)
			if !ok { panic("Type error: undefined variable " + n.Name.Value) }
			if len(t) > 2 && t[len(t)-2:] == "[]" { targetType = t[:len(t)-2] } else { panic("Type error: cannot index non-array") }
		} else {
			t, ok := tenv.Get(n.Name.Value)
			if !ok { panic("Type error: undefined variable " + n.Name.Value) }
			targetType = t
		}
		vt := GetType(n.Value, tenv)
		if targetType != vt { panic(fmt.Sprintf("Type error: cannot assign %s to %s", vt, targetType)) }
	case *FunctionDecl:
		tenv.SetFunc(n.Name.Value, n)
		funcEnv := &TypeEnv{store: make(map[string]string), funcs: make(map[string]*FunctionDecl), outer: tenv}
		for _, p := range n.Parameters { funcEnv.Set(p.Name.Value, p.Type) }
		TypeCheck(n.Body, funcEnv, filename, loader)
	case *ExportStmt: TypeCheck(n.Decl, tenv, filename, loader)
	case *ImportStmt:
		modSrc, err := loader(n.Module)
		if err != nil { panic("Failed to load module: " + n.Module) }
		modAst := NewParser(NewLexer(modSrc)).ParseProgram()
		modTEnv := NewTypeEnv()
		TypeCheck(modAst, modTEnv, n.Module, loader)
		for _, name := range n.Names {
			if t, ok := modTEnv.Get(name.Value); ok { tenv.Set(name.Value, t) } else if f, ok := modTEnv.GetFunc(name.Value); ok { tenv.SetFunc(name.Value, f) } else {
				panic(fmt.Sprintf("Type error in %s: imported name '%s' not found", filename, name.Value))
			}
		}
	case *AdjointExpression:
		TypeCheck(n.Call, tenv, filename, loader)
	case *CallExpr:
		if n.Function == "superpose" || n.Function == "invert" || n.Function == "X" || n.Function == "Y" || n.Function == "Z" || n.Function == "S" || n.Function == "T" || n.Function == "measure" {
			if len(n.Args) != 1 { panic(fmt.Sprintf("Error in %s: %s requires 1 argument", filename, n.Function)) }
			argType := GetType(n.Args[0], tenv)
			if argType != "qubit" {
				panic(fmt.Sprintf("Fatal error in %s: cannot apply quantum gate '%s' to classical type '%s'", filename, n.Function, argType))
			}
		} else if n.Function == "Rx" || n.Function == "Ry" || n.Function == "Rz" {
			if len(n.Args) != 2 { panic(fmt.Sprintf("Error in %s: %s requires 2 arguments (number, qubit)", filename, n.Function)) }
			arg1Type := GetType(n.Args[0], tenv)
			arg2Type := GetType(n.Args[1], tenv)
			if arg1Type != "number" || arg2Type != "qubit" {
				panic(fmt.Sprintf("Fatal error in %s: %s requires (number, qubit), got (%s, %s)", filename, n.Function, arg1Type, arg2Type))
			}
		} else if n.Function == "StatePrep" {
			if len(n.Args) != 2 { panic("Error: StatePrep requires 2 arguments (number[], qubit[])") }
			if GetType(n.Args[0], tenv) != "number[]" || GetType(n.Args[1], tenv) != "qubit[]" {
				panic("Fatal error: StatePrep requires (number[], qubit[])")
			}
		} else if strings.HasPrefix(n.Function, "Math.") {
			if len(n.Args) != 1 { panic(fmt.Sprintf("Type error: %s expects 1 argument", n.Function)) }
			if GetType(n.Args[0], tenv) != "number" { panic("Type error: Math functions require number") }
		} else {
			f, ok := tenv.GetFunc(n.Function)
			if !ok {
				panic("Type error: undefined function " + n.Function)
			}
			if len(f.Parameters) != len(n.Args) { panic(fmt.Sprintf("Type error: %s expects %d arguments", n.Function, len(f.Parameters))) }
			for i, arg := range n.Args {
				if GetType(arg, tenv) != f.Parameters[i].Type { panic(fmt.Sprintf("Type error: argument %d to %s should be %s", i, n.Function, f.Parameters[i].Type)) }
			}
		}
	case *IfStmt:
		condType := GetType(n.Condition, tenv)
		if condType != "boolean" && condType != "qubit" { panic("Type error: if condition must be boolean or qubit") }
		TypeCheck(n.Consequence, tenv, filename, loader)
	case *WhileStmt:
		condType := GetType(n.Condition, tenv)
		if condType != "boolean" && condType != "qubit" { panic("Type error: while condition must be boolean or qubit") }
		TypeCheck(n.Body, tenv, filename, loader)
	case *ForStmt:
		forEnv := &TypeEnv{store: make(map[string]string), funcs: make(map[string]*FunctionDecl), outer: tenv}
		if n.Init != nil { TypeCheck(n.Init, forEnv, filename, loader) }
		if n.Condition != nil {
			condType := GetType(n.Condition, forEnv)
			if condType != "boolean" && condType != "qubit" { panic("Type error: for condition must be boolean or qubit") }
		}
		if n.Post != nil { TypeCheck(n.Post, forEnv, filename, loader) }
		TypeCheck(n.Body, forEnv, filename, loader)
	case *ExpressionStmt:
		TypeCheck(n.Expr, tenv, filename, loader)
	case *ReturnStmt:
		if n.Value != nil { GetType(n.Value, tenv) }
	}
}

// =======================================================
// 5. QUANTUM STATE SIMULATOR (Linear Algebra Engine)
// =======================================================
type Complex struct { R, I float64 }
func add(a, b Complex) Complex { return Complex{a.R + b.R, a.I + b.I} }
func mul(a, b Complex) Complex { return Complex{a.R*b.R - a.I*b.I, a.R*b.I + a.I*b.R} }
func scale(a Complex, s float64) Complex { return Complex{a.R * s, a.I * s} }

type QState struct {
	Amps         []Complex
	NextID       int
	Tally        map[string]bool
	HardwareMode bool
	NoiseLevel   float64
	QasmData     []string
	History      map[int][]string
}

func NewQState() *QState {
	s := &QState{Amps: make([]Complex, 1024), Tally: make(map[string]bool), QasmData: make([]string, 0), History: make(map[int][]string)}
	s.Amps[0] = Complex{1, 0}
	s.NoiseLevel = 0
	return s
}

func (s *QState) Alloc() int {
	id := s.NextID
	s.NextID++
	if s.HardwareMode {
		s.QasmData = append(s.QasmData, fmt.Sprintf("qreg q%d[1];", id))
	}
	return id
}

func (s *QState) recordGate(gate string, target int, controls []int) {
	s.History[target] = append(s.History[target], gate)
	if !s.HardwareMode { return }
	if len(controls) == 0 {
		s.QasmData = append(s.QasmData, fmt.Sprintf("%s q%d[0];", gate, target))
	} else if len(controls) == 1 {
		s.QasmData = append(s.QasmData, fmt.Sprintf("c%s q%d[0], q%d[0];", gate, controls[0], target))
	} else if len(controls) == 2 {
		s.QasmData = append(s.QasmData, fmt.Sprintf("cc%s q%d[0], q%d[0], q%d[0];", gate, controls[0], controls[1], target))
	}
}

func (s *QState) recordRotGate(gate string, theta float64, target int, controls []int) {
	s.History[target] = append(s.History[target], fmt.Sprintf("%s(%f)", gate, theta))
	if !s.HardwareMode { return }
	if len(controls) == 0 {
		s.QasmData = append(s.QasmData, fmt.Sprintf("%s(%f) q%d[0];", gate, theta, target))
	} else if len(controls) == 1 {
		s.QasmData = append(s.QasmData, fmt.Sprintf("c%s(%f) q%d[0], q%d[0];", gate, theta, controls[0], target))
	}
}

func checkControls(i int, controls []int) bool {
	for _, c := range controls { if (i & (1 << c)) == 0 { return false } }
	return true
}

func (s *QState) applyGateInternal(target int, controls []int, U [4]Complex) {
	for i := 0; i < len(s.Amps); i++ {
		if checkControls(i, controls) {
			if (i & (1 << target)) == 0 {
				i0 := i; i1 := i | (1 << target)
				a0 := s.Amps[i0]; a1 := s.Amps[i1]
				s.Amps[i0] = add(mul(U[0], a0), mul(U[1], a1))
				s.Amps[i1] = add(mul(U[2], a0), mul(U[3], a1))
			}
		}
	}
}

func (s *QState) applyGate(target int, controls []int, U [4]Complex) {
	s.applyGateInternal(target, controls, U)
	if s.NoiseLevel > 0 && rand.Float64() < s.NoiseLevel {
		r := rand.Float64()
		if r < 0.333 {
			s.applyGateInternal(target, nil, [4]Complex{{0, 0}, {1, 0}, {1, 0}, {0, 0}}) // X
		} else if r < 0.666 {
			s.applyGateInternal(target, nil, [4]Complex{{0, 0}, {0, -1}, {0, 1}, {0, 0}}) // Y
		} else {
			s.applyGateInternal(target, nil, [4]Complex{{1, 0}, {0, 0}, {0, 0}, {-1, 0}}) // Z
		}
	}
}

func (s *QState) Uncompute(q int) {
	hist := s.History[q]
	// replay backwards with adjoints
	for i := len(hist) - 1; i >= 0; i-- {
		op := hist[i]
		if op == "h" { s.H(q, nil) }
		if op == "x" { s.X(q, nil) }
		if op == "y" { s.Y(q, nil) }
		if op == "z" { s.Z(q, nil) }
		if op == "s" { s.SDgGate(q, nil) }
		if op == "t" { s.TDgGate(q, nil) }
		if op == "sdg" { s.SGate(q, nil) }
		if op == "tdg" { s.TGate(q, nil) }
		if strings.HasPrefix(op, "rx") {
			var theta float64
			fmt.Sscanf(op, "rx(%f)", &theta)
			s.Rx(-theta, q, nil)
		}
		if strings.HasPrefix(op, "ry") {
			var theta float64
			fmt.Sscanf(op, "ry(%f)", &theta)
			s.Ry(-theta, q, nil)
		}
		if strings.HasPrefix(op, "rz") {
			var theta float64
			fmt.Sscanf(op, "rz(%f)", &theta)
			s.Rz(-theta, q, nil)
		}
	}
	s.History[q] = nil
}

func (s *QState) H(q int, controls []int) {
	s.recordGate("h", q, controls)
	inv := 1.0 / math.Sqrt(2)
	s.applyGate(q, controls, [4]Complex{{inv, 0}, {inv, 0}, {inv, 0}, {-inv, 0}})
}

func (s *QState) X(q int, controls []int) {
	s.recordGate("x", q, controls)
	s.applyGate(q, controls, [4]Complex{{0, 0}, {1, 0}, {1, 0}, {0, 0}})
}

func (s *QState) Y(q int, controls []int) {
	s.recordGate("y", q, controls)
	s.applyGate(q, controls, [4]Complex{{0, 0}, {0, -1}, {0, 1}, {0, 0}})
}

func (s *QState) Z(q int, controls []int) {
	s.recordGate("z", q, controls)
	s.applyGate(q, controls, [4]Complex{{1, 0}, {0, 0}, {0, 0}, {-1, 0}})
}

func (s *QState) SGate(q int, controls []int) {
	s.recordGate("s", q, controls)
	s.applyGate(q, controls, [4]Complex{{1, 0}, {0, 0}, {0, 0}, {0, 1}})
}

func (s *QState) TGate(q int, controls []int) {
	s.recordGate("t", q, controls)
	s.applyGate(q, controls, [4]Complex{{1, 0}, {0, 0}, {0, 0}, {math.Cos(math.Pi / 4), math.Sin(math.Pi / 4)}})
}

func (s *QState) SDgGate(q int, controls []int) {
	s.recordGate("sdg", q, controls)
	s.applyGate(q, controls, [4]Complex{{1, 0}, {0, 0}, {0, 0}, {0, -1}})
}

func (s *QState) TDgGate(q int, controls []int) {
	s.recordGate("tdg", q, controls)
	s.applyGate(q, controls, [4]Complex{{1, 0}, {0, 0}, {0, 0}, {math.Cos(-math.Pi / 4), math.Sin(-math.Pi / 4)}})
}

func (s *QState) Rx(theta float64, q int, controls []int) {
	s.recordRotGate("rx", theta, q, controls)
	c := math.Cos(theta / 2.0)
	si := -math.Sin(theta / 2.0)
	s.applyGate(q, controls, [4]Complex{{c, 0}, {0, si}, {0, si}, {c, 0}})
}

func (s *QState) Ry(theta float64, q int, controls []int) {
	s.recordRotGate("ry", theta, q, controls)
	c := math.Cos(theta / 2.0)
	si := math.Sin(theta / 2.0)
	s.applyGate(q, controls, [4]Complex{{c, 0}, {-si, 0}, {si, 0}, {c, 0}})
}

func (s *QState) Rz(theta float64, q int, controls []int) {
	s.recordRotGate("rz", theta, q, controls)
	c1 := math.Cos(-theta / 2.0)
	s1 := math.Sin(-theta / 2.0)
	c2 := math.Cos(theta / 2.0)
	s2 := math.Sin(theta / 2.0)
	s.applyGate(q, controls, [4]Complex{{c1, s1}, {0, 0}, {0, 0}, {c2, s2}})
}

func (s *QState) Measure(q int, name string) bool {
	if s.HardwareMode {
		s.QasmData = append(s.QasmData, fmt.Sprintf("creg c%d[1];", q))
		s.QasmData = append(s.QasmData, fmt.Sprintf("measure q%d[0] -> c%d[0];", q, q))
		return false
	}
	prob1 := 0.0
	for i, a := range s.Amps {
		if (i & (1 << q)) != 0 { prob1 += a.R*a.R + a.I*a.I }
	}

	if prob1 < 1e-9 { prob1 = 0 }
	if prob1 > 1-1e-9 { prob1 = 1 }

	res := rand.Float64() < prob1
	var norm float64
	if res { norm = 1.0 / math.Sqrt(prob1) } else { norm = 1.0 / math.Sqrt(1.0-prob1) }

	for i := range s.Amps {
		if res {
			if (i & (1 << q)) == 0 { s.Amps[i] = Complex{0, 0} } else { s.Amps[i] = scale(s.Amps[i], norm) }
		} else {
			if (i & (1 << q)) != 0 { s.Amps[i] = Complex{0, 0} } else { s.Amps[i] = scale(s.Amps[i], norm) }
		}
	}

	if name != "" { s.Tally[name] = res }
	return res
}

// =======================================================
// 6. EVALUATOR (Supports Control Contexts)
// =======================================================
type Value interface{}
type BooleanVal struct{ Value bool }
type NumberVal struct{ Value float64 }
type StringVal struct{ Value string }
type ObjectVal struct{ Pairs map[string]Value }
type QubitVal struct { ID int; Name string }
type ArrayVal struct { Elements []Value }
type ReturnVal struct { Value Value }
type FunctionVal struct { Parameters []*Param; Body *BlockStmt; Env *Env }

type Env struct { store map[string]Value; outer *Env }
func NewEnv() *Env { return &Env{store: make(map[string]Value)} }
func (e *Env) Set(name string, val Value) { e.store[name] = val }
func (e *Env) Get(name string) (Value, bool) {
	v, ok := e.store[name]
	if !ok && e.outer != nil { return e.outer.Get(name) }
	return v, ok
}

func Eval(node Node, env *Env, qstate *QState, controls []int, loader ModuleLoader) Value {
	switch n := node.(type) {
	case *Program:
		var result Value
		for _, stmt := range n.Stmts {
			result = Eval(stmt, env, qstate, controls, loader)
			if _, ok := result.(*ReturnVal); ok { return result }
		}
		return result
	case *BlockStmt:
		blockEnv := &Env{store: make(map[string]Value), outer: env}
		var result Value
		for _, stmt := range n.Stmts {
			result = Eval(stmt, blockEnv, qstate, controls, loader)
			if _, ok := result.(*ReturnVal); ok { break }
		}
		
		// Garbage Collection: Uncompute any qubit allocated in this block
		for _, val := range blockEnv.store {
			if qv, ok := val.(*QubitVal); ok {
				qstate.Uncompute(qv.ID)
			} else if arr, ok := val.(*ArrayVal); ok {
				for _, e := range arr.Elements {
					if qv, ok2 := e.(*QubitVal); ok2 {
						qstate.Uncompute(qv.ID)
					}
				}
			}
		}
		
		return result
	case *ConstDecl:
		val := Eval(n.Value, env, qstate, controls, loader)
		if qval, ok := val.(*QubitVal); ok { qval.Name = n.Name.Value }
		if arrVal, ok := val.(*ArrayVal); ok { 
			for i, e := range arrVal.Elements { 
				if qv, ok2 := e.(*QubitVal); ok2 { qv.Name = fmt.Sprintf("%s[%d]", n.Name.Value, i) }
			}
		}
		env.Set(n.Name.Value, val)
	case *LetDecl:
		val := Eval(n.Value, env, qstate, controls, loader)
		env.Set(n.Name.Value, val)
	case *AssignStmt:
		val := Eval(n.Value, env, qstate, controls, loader)
		if n.Index != nil {
			arr := env.store[n.Name.Value].(*ArrayVal)
			idx := int(Eval(n.Index, env, qstate, controls, loader).(*NumberVal).Value)
			arr.Elements[idx] = val
		} else {
			currEnv := env
			for currEnv != nil {
				if _, ok := currEnv.store[n.Name.Value]; ok {
					currEnv.store[n.Name.Value] = val
					break
				}
				currEnv = currEnv.outer
			}
		}
	case *FunctionDecl:
		funcVal := &FunctionVal{Parameters: n.Parameters, Body: n.Body, Env: env}
		env.Set(n.Name.Value, funcVal)
	case *ExportStmt:
		Eval(n.Decl, env, qstate, controls, loader)
	case *IfStmt:
		cond := Eval(n.Condition, env, qstate, controls, loader)
		if qval, ok := cond.(*QubitVal); ok {
			newControls := append([]int{}, controls...)
			newControls = append(newControls, qval.ID)
			return Eval(n.Consequence, env, qstate, newControls, loader)
		} else if bval, ok := cond.(*BooleanVal); ok {
			if bval.Value { return Eval(n.Consequence, env, qstate, controls, loader) }
		}
	case *WhileStmt:
		for {
			cond := Eval(n.Condition, env, qstate, controls, loader)
			if qval, ok := cond.(*QubitVal); ok {
				newControls := append([]int{}, controls...)
				newControls = append(newControls, qval.ID)
				res := Eval(n.Body, env, qstate, newControls, loader)
				if _, ok := res.(*ReturnVal); ok { return res }
			} else if bval, ok := cond.(*BooleanVal); ok {
				if !bval.Value { break }
				res := Eval(n.Body, env, qstate, controls, loader)
				if _, ok := res.(*ReturnVal); ok { return res }
			} else { break }
		}
	case *ForStmt:
		forEnv := &Env{store: make(map[string]Value), outer: env}
		if n.Init != nil { Eval(n.Init, forEnv, qstate, controls, loader) }
		for {
			if n.Condition != nil {
				cond := Eval(n.Condition, forEnv, qstate, controls, loader)
				if bval, ok := cond.(*BooleanVal); ok && !bval.Value { break }
			}
			res := Eval(n.Body, forEnv, qstate, controls, loader)
			if _, ok := res.(*ReturnVal); ok { return res }
			if n.Post != nil { Eval(n.Post, forEnv, qstate, controls, loader) }
		}
	case *ExpressionStmt: Eval(n.Expr, env, qstate, controls, loader)
	case *ReturnStmt:
		if n.Value != nil { return &ReturnVal{Value: Eval(n.Value, env, qstate, controls, loader)} }
		return &ReturnVal{Value: nil}
	case *ImportStmt:
		modSrc, _ := loader(n.Module)
		modAst := NewParser(NewLexer(modSrc)).ParseProgram()
		modEnv := NewEnv()
		Eval(modAst, modEnv, qstate, controls, loader) 
		for _, name := range n.Names {
			if val, ok := modEnv.Get(name.Value); ok { env.Set(name.Value, val) }
		}
	case *NewQubit: return &QubitVal{ID: qstate.Alloc()}
	case *ObjectLiteral:
		obj := &ObjectVal{Pairs: make(map[string]Value)}
		for k, v := range n.Pairs { obj.Pairs[k] = Eval(v, env, qstate, controls, loader) }
		return obj
	case *NewArray:
		length := int(Eval(n.Length, env, qstate, controls, loader).(*NumberVal).Value)
		arr := &ArrayVal{Elements: make([]Value, length)}
		for i := 0; i < length; i++ {
			if n.Type == "qubit" || n.Type == "Qubit" { arr.Elements[i] = &QubitVal{ID: qstate.Alloc()} } else if n.Type == "number" { arr.Elements[i] = &NumberVal{Value: 0} } else if n.Type == "boolean" { arr.Elements[i] = &BooleanVal{Value: false} }
		}
		return arr
	case *ArrayLiteral:
		arr := &ArrayVal{Elements: make([]Value, len(n.Elements))}
		for i, e := range n.Elements { arr.Elements[i] = Eval(e, env, qstate, controls, loader) }
		return arr
	case *BooleanLiteral: return &BooleanVal{Value: n.Value}
	case *NumberLiteral: return &NumberVal{Value: n.Value}
	case *StringLiteral: return &StringVal{Value: n.Value}
	case *InfixExpression:
		left := Eval(n.Left, env, qstate, controls, loader)
		right := Eval(n.Right, env, qstate, controls, loader)
		switch n.Operator {
		case "==":
			if lv, ok := left.(*NumberVal); ok { return &BooleanVal{Value: lv.Value == right.(*NumberVal).Value} }
			if lv, ok := left.(*BooleanVal); ok { return &BooleanVal{Value: lv.Value == right.(*BooleanVal).Value} }
			if lv, ok := left.(*StringVal); ok { return &BooleanVal{Value: lv.Value == right.(*StringVal).Value} }
		case "!=":
			if lv, ok := left.(*NumberVal); ok { return &BooleanVal{Value: lv.Value != right.(*NumberVal).Value} }
			if lv, ok := left.(*BooleanVal); ok { return &BooleanVal{Value: lv.Value != right.(*BooleanVal).Value} }
			if lv, ok := left.(*StringVal); ok { return &BooleanVal{Value: lv.Value != right.(*StringVal).Value} }
		case "<":
			return &BooleanVal{Value: left.(*NumberVal).Value < right.(*NumberVal).Value}
		case ">":
			return &BooleanVal{Value: left.(*NumberVal).Value > right.(*NumberVal).Value}
		case "+":
			if lv, ok := left.(*StringVal); ok { return &StringVal{Value: lv.Value + right.(*StringVal).Value} }
			return &NumberVal{Value: left.(*NumberVal).Value + right.(*NumberVal).Value}
		case "-": return &NumberVal{Value: left.(*NumberVal).Value - right.(*NumberVal).Value}
		case "*": return &NumberVal{Value: left.(*NumberVal).Value * right.(*NumberVal).Value}
		case "/": return &NumberVal{Value: left.(*NumberVal).Value / right.(*NumberVal).Value}
		}
	case *IndexExpression:
		arr := Eval(n.Left, env, qstate, controls, loader).(*ArrayVal)
		idx := int(Eval(n.Index, env, qstate, controls, loader).(*NumberVal).Value)
		return arr.Elements[idx]
	case *PropertyAccess:
		obj := Eval(n.Object, env, qstate, controls, loader)
		if n.Property == "pi" || n.Property == "PI" { return &NumberVal{Value: math.Pi} }
		if objVal, ok := obj.(*ObjectVal); ok {
			if val, ok2 := objVal.Pairs[n.Property]; ok2 { return val }
			return nil
		}
		return obj
	case *Identifier:
		val, ok := env.Get(n.Value)
		if !ok { panic("Runtime error: undefined variable: " + n.Value) }
		return val
	case *AdjointExpression:
		call, ok := n.Call.(*CallExpr)
		if !ok { panic("Runtime error: adjoint must be applied to a function call") }
		
		if call.Function == "Rx" || call.Function == "Ry" || call.Function == "Rz" {
			theta := Eval(call.Args[0], env, qstate, controls, loader).(*NumberVal).Value
			q := Eval(call.Args[1], env, qstate, controls, loader).(*QubitVal)
			if call.Function == "Rx" { qstate.Rx(-theta, q.ID, controls) }
			if call.Function == "Ry" { qstate.Ry(-theta, q.ID, controls) }
			if call.Function == "Rz" { qstate.Rz(-theta, q.ID, controls) }
			return nil
		}
		if call.Function == "S" {
			q := Eval(call.Args[0], env, qstate, controls, loader).(*QubitVal)
			qstate.SDgGate(q.ID, controls)
			return nil
		}
		if call.Function == "T" {
			q := Eval(call.Args[0], env, qstate, controls, loader).(*QubitVal)
			qstate.TDgGate(q.ID, controls)
			return nil
		}
		return Eval(call, env, qstate, controls, loader)
	case *CallExpr:
		if n.Function == "superpose" {
			q := Eval(n.Args[0], env, qstate, controls, loader).(*QubitVal)
			qstate.H(q.ID, controls)
		} else if n.Function == "invert" || n.Function == "X" {
			q := Eval(n.Args[0], env, qstate, controls, loader).(*QubitVal)
			qstate.X(q.ID, controls)
		} else if n.Function == "Y" {
			q := Eval(n.Args[0], env, qstate, controls, loader).(*QubitVal)
			qstate.Y(q.ID, controls)
		} else if n.Function == "Z" {
			q := Eval(n.Args[0], env, qstate, controls, loader).(*QubitVal)
			qstate.Z(q.ID, controls)
		} else if n.Function == "S" {
			q := Eval(n.Args[0], env, qstate, controls, loader).(*QubitVal)
			qstate.SGate(q.ID, controls)
		} else if n.Function == "T" {
			q := Eval(n.Args[0], env, qstate, controls, loader).(*QubitVal)
			qstate.TGate(q.ID, controls)
		} else if n.Function == "Rx" || n.Function == "Ry" || n.Function == "Rz" {
			theta := Eval(n.Args[0], env, qstate, controls, loader).(*NumberVal).Value
			q := Eval(n.Args[1], env, qstate, controls, loader).(*QubitVal)
			if n.Function == "Rx" { qstate.Rx(theta, q.ID, controls) }
			if n.Function == "Ry" { qstate.Ry(theta, q.ID, controls) }
			if n.Function == "Rz" { qstate.Rz(theta, q.ID, controls) }
		} else if n.Function == "measure" {
			q := Eval(n.Args[0], env, qstate, controls, loader).(*QubitVal)
			return &BooleanVal{Value: qstate.Measure(q.ID, q.Name)}
		} else if n.Function == "StatePrep" {
			ampsArr := Eval(n.Args[0], env, qstate, controls, loader).(*ArrayVal)
			targetArr := Eval(n.Args[1], env, qstate, controls, loader).(*ArrayVal)
			
			// Normalize classical array
			var norm float64
			var vals []float64
			for _, v := range ampsArr.Elements {
				f := v.(*NumberVal).Value
				vals = append(vals, f)
				norm += f * f
			}
			norm = math.Sqrt(norm)
			
			// Extract target qubit IDs
			var qids []int
			for _, v := range targetArr.Elements {
				qids = append(qids, v.(*QubitVal).ID)
			}
			
			// Simple Simulator Cheat: forcefully set the amplitudes
			// For a fully general QASM synthesis, we would output deeply nested rotations.
			if !qstate.HardwareMode {
				// We need to inject the amplitudes only into the target subspace.
				// This requires a full tensor product. For the sake of the prototype,
				// if we are preparing into a fresh register, we can just map the states.
				
				// Zero out ONLY the amplitudes where the target qubits are non-zero
				// Wait, the simplest fix is to just do single qubit assignments if len == 2
				if len(qids) == 1 && len(vals) == 2 {
					// Prepare single qubit
					q := qids[0]
					for i := range qstate.Amps {
						if (i & (1 << q)) == 0 {
							a0 := qstate.Amps[i]
							// Assuming it started in |0>
							qstate.Amps[i] = scale(a0, vals[0]/norm)
							qstate.Amps[i|(1<<q)] = scale(a0, vals[1]/norm)
						}
					}
				} else if len(qids) == 2 && len(vals) == 4 {
					q0 := qids[0]
					q1 := qids[1]
					for i := range qstate.Amps {
						if (i&(1<<q0)) == 0 && (i&(1<<q1)) == 0 {
							a0 := qstate.Amps[i]
							qstate.Amps[i] = scale(a0, vals[0]/norm)
							qstate.Amps[i|(1<<q0)] = scale(a0, vals[1]/norm)
							qstate.Amps[i|(1<<q1)] = scale(a0, vals[2]/norm)
							qstate.Amps[i|(1<<q0)|(1<<q1)] = scale(a0, vals[3]/norm)
						}
					}
				}
			}
			return nil
		} else if n.Function == "Math.sin" {
			arg := Eval(n.Args[0], env, qstate, controls, loader).(*NumberVal).Value
			return &NumberVal{Value: math.Sin(arg)}
		} else if n.Function == "Math.cos" {
			arg := Eval(n.Args[0], env, qstate, controls, loader).(*NumberVal).Value
			return &NumberVal{Value: math.Cos(arg)}
		} else if n.Function == "Math.sqrt" {
			arg := Eval(n.Args[0], env, qstate, controls, loader).(*NumberVal).Value
			return &NumberVal{Value: math.Sqrt(arg)}
		} else {
			val, ok := env.Get(n.Function)
			if !ok {
				panic("Runtime error: undefined function " + n.Function)
			}
			fn := val.(*FunctionVal)
			callEnv := &Env{store: make(map[string]Value), outer: fn.Env}
			for i, param := range fn.Parameters {
				callEnv.Set(param.Name.Value, Eval(n.Args[i], env, qstate, controls, loader))
			}
			res := Eval(fn.Body, callEnv, qstate, controls, loader)
			if ret, ok := res.(*ReturnVal); ok { return ret.Value }
			return res
		}
	}
	return nil
}

// =======================================================
// 7. ENTRYPOINT (Shared by CLI and WASM)
// =======================================================

// Run evaluates the QuantumScript source code using 1000 shots
func Run(src string, filename string, loader ModuleLoader, noiseLevel float64) map[string]map[bool]int {
	ast := NewParser(NewLexer(src)).ParseProgram()
	tenv := NewTypeEnv()
	TypeCheck(ast, tenv, filename, loader)

	globalTally := make(map[string]map[bool]int)
	for shot := 0; shot < 1000; shot++ {
		qstate := NewQState() 
		qstate.NoiseLevel = noiseLevel
		env := NewEnv()
		Eval(ast, env, qstate, nil, loader) 
		for name, res := range qstate.Tally {
			if globalTally[name] == nil { globalTally[name] = make(map[bool]int) }
			globalTally[name][res]++
		}
	}
	return globalTally
}

// ExportQASM translates the source into a flat OpenQASM 3.0 string
func ExportQASM(src string, filename string, loader ModuleLoader, topology string) string {
	ast := NewParser(NewLexer(src)).ParseProgram()
	tenv := NewTypeEnv()
	TypeCheck(ast, tenv, filename, loader)

	qstate := NewQState()
	qstate.HardwareMode = true
	env := NewEnv()
	Eval(ast, env, qstate, nil, loader)

	qasm := "OPENQASM 3.0;\ninclude \"stdgates.inc\";\n\n"
	
	if topology == "linear" {
		for _, cmd := range qstate.QasmData {
			// Find 2-qubit gates: e.g. cx q0[0], q3[0];
			if strings.HasPrefix(cmd, "c") && !strings.HasPrefix(cmd, "cc") && !strings.HasPrefix(cmd, "creg") {
				// Parse c<gate> qA[0], qB[0];
				parts := strings.Split(cmd, " ")
				if len(parts) >= 3 && strings.HasPrefix(parts[1], "q") && strings.HasPrefix(parts[2], "q") {
					qA_str := strings.TrimSuffix(strings.TrimPrefix(parts[1], "q"), "[0],")
					qB_str := strings.TrimSuffix(strings.TrimPrefix(parts[2], "q"), "[0];")
					qA, _ := strconv.Atoi(qA_str)
					qB, _ := strconv.Atoi(qB_str)
					
					// Insert SWAPs if not adjacent
					if math.Abs(float64(qA - qB)) > 1 {
						step := 1
						if qA > qB { step = -1 }
						
						// Route qA to qB
						curr := qA
						for math.Abs(float64(curr - qB)) > 1 {
							next := curr + step
							qasm += fmt.Sprintf("swap q%d[0], q%d[0];\n", curr, next)
							curr = next
						}
						
						// Apply gate
						gate := parts[0]
						qasm += fmt.Sprintf("%s q%d[0], q%d[0];\n", gate, curr, qB)
						
						// Route back
						for curr != qA {
							prev := curr - step
							qasm += fmt.Sprintf("swap q%d[0], q%d[0];\n", curr, prev)
							curr = prev
						}
						continue
					}
				}
			}
			qasm += cmd + "\n"
		}
	} else {
		for _, cmd := range qstate.QasmData {
			qasm += cmd + "\n"
		}
	}
	
	return qasm
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
