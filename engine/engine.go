package engine

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// =======================================================
// 1. LEXER
// =======================================================
type TokenType string

const (
	CONST TokenType = "CONST"; BOOLEAN TokenType = "BOOLEAN"; QUBIT TokenType = "QUBIT"
	NEW TokenType = "NEW"; TRUE TokenType = "TRUE"; FALSE TokenType = "FALSE"
	SUPERPOSE TokenType = "SUPERPOSE"; INVERT TokenType = "INVERT"; MEASURE TokenType = "MEASURE"
	IF TokenType = "IF"; IMPORT TokenType = "IMPORT"; FROM TokenType = "FROM"; EXPORT TokenType = "EXPORT"
	IDENT TokenType = "IDENT"; LBRACE TokenType = "LBRACE"; RBRACE TokenType = "RBRACE"
	LPAREN TokenType = "LPAREN"; RPAREN TokenType = "RPAREN"; ASSIGN TokenType = "ASSIGN"
	SEMI TokenType = "SEMI"; COLON TokenType = "COLON"; STRING TokenType = "STRING"
	COMMA TokenType = "COMMA"; EOF TokenType = "EOF"; ILLEGAL TokenType = "ILLEGAL"
)

var keywords = map[string]TokenType{
	"const": CONST, "boolean": BOOLEAN, "qubit": QUBIT, "new": NEW, "true": TRUE, "false": FALSE,
	"superpose": SUPERPOSE, "invert": INVERT, "measure": MEASURE, "if": IF,
	"import": IMPORT, "from": FROM, "export": EXPORT,
}

type Token struct { Type TokenType; Literal string }

type Lexer struct { input string; pos int; readPos int; ch byte }

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
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

func (l *Lexer) readIdentifier() string {
	pos := l.pos
	for isLetter(l.ch) || ('0' <= l.ch && l.ch <= '9') { l.readChar() }
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

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()
	var tok Token
	switch l.ch {
	case '=': tok = Token{ASSIGN, "="}
	case ';': tok = Token{SEMI, ";"}
	case ':': tok = Token{COLON, ":"}
	case '(': tok = Token{LPAREN, "("}
	case ')': tok = Token{RPAREN, ")"}
	case '{': tok = Token{LBRACE, "{"}
	case '}': tok = Token{RBRACE, "}"}
	case ',': tok = Token{COMMA, ","}
	case '"': tok = Token{STRING, l.readString()}
	case '/':
		if l.peekChar() == '/' {
			l.skipComment()
			return l.NextToken()
		}
		tok = Token{ILLEGAL, string(l.ch)}
	case 0: tok = Token{EOF, ""}
	default:
		if isLetter(l.ch) {
			ident := l.readIdentifier()
			if t, ok := keywords[ident]; ok { tok.Type = t } else { tok.Type = IDENT }
			tok.Literal = ident
			return tok
		}
		tok = Token{ILLEGAL, string(l.ch)}
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
type ExportStmt struct{ Decl *ConstDecl }
type IfStmt struct { Condition Expression; Consequence *BlockStmt }
type ImportStmt struct { Names []*Identifier; Module string }
type BlockStmt struct{ Stmts []Statement }
type ExpressionStmt struct{ Expr Expression }

type Identifier struct{ Value string }
type BooleanLiteral struct{ Value bool }
type StringLiteral struct{ Value string }
type NewQubit struct{}
type CallExpr struct { Function string; Args []Expression }

func (p *Program) stmt() {} 
func (c *ConstDecl) stmt() {} 
func (i *IfStmt) stmt() {}
func (i *ImportStmt) stmt() {} 
func (e *ExportStmt) stmt() {} 
func (e *ExpressionStmt) stmt() {}
func (b *BlockStmt) stmt() {} 
func (i *Identifier) expr() {} 
func (b *BooleanLiteral) expr() {}
func (s *StringLiteral) expr() {} 
func (n *NewQubit) expr() {} 
func (c *CallExpr) expr() {}

// =======================================================
// 3. PARSER
// =======================================================
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
	} else { panic(fmt.Sprintf("Syntax error: expected %s got %s", t, p.peekToken.Type)) }
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
	case EXPORT: return p.parseExportStmt()
	case IF: return p.parseIfStmt()
	case IMPORT: return p.parseImportStmt()
	default: return p.parseExpressionStmt()
	}
}

func (p *Parser) parseConstDecl() *ConstDecl {
	decl := &ConstDecl{}
	p.expectPeek(IDENT)
	decl.Name = &Identifier{Value: p.curToken.Literal}
	p.expectPeek(COLON)
	p.nextToken() 
	if p.curToken.Type != BOOLEAN && p.curToken.Type != QUBIT { panic("Syntax error: expected boolean or qubit type") }
	decl.Type = p.curToken.Literal
	p.expectPeek(ASSIGN)
	p.nextToken() 
	decl.Value = p.parseExpression()
	p.expectPeek(SEMI)
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

func (p *Parser) parseIfStmt() *IfStmt {
	stmt := &IfStmt{}
	p.expectPeek(LPAREN)
	p.nextToken()
	stmt.Condition = p.parseExpression()
	p.expectPeek(RPAREN)
	p.expectPeek(LBRACE)
	stmt.Consequence = &BlockStmt{}
	p.nextToken() 
	for p.curToken.Type != RBRACE && p.curToken.Type != EOF {
		if s := p.parseStatement(); s != nil { stmt.Consequence.Stmts = append(stmt.Consequence.Stmts, s) }
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpressionStmt() *ExpressionStmt {
	stmt := &ExpressionStmt{Expr: p.parseExpression()}
	p.expectPeek(SEMI)
	return stmt
}

func (p *Parser) parseExpression() Expression {
	var left Expression
	switch p.curToken.Type {
	case NEW:
		p.expectPeek(IDENT)
		if p.curToken.Literal != "Qubit" { panic("Syntax error: expected Qubit") }
		p.expectPeek(LPAREN); p.expectPeek(RPAREN)
		left = &NewQubit{}
	case TRUE: left = &BooleanLiteral{Value: true}
	case FALSE: left = &BooleanLiteral{Value: false}
	case IDENT, SUPERPOSE, INVERT, MEASURE:
		left = &Identifier{Value: p.curToken.Literal}
	}

	if p.peekToken.Type == LPAREN {
		p.nextToken() 
		call := &CallExpr{Function: left.(*Identifier).Value}
		p.nextToken() 
		if p.curToken.Type != RPAREN {
			for {
				call.Args = append(call.Args, p.parseExpression())
				if p.peekToken.Type == COMMA { p.nextToken(); p.nextToken() } else { break }
			}
		}
		p.expectPeek(RPAREN)
		left = call
	}
	return left
}

// =======================================================
// 4. STATIC SEMANTIC TYPE CHECKER
// =======================================================
type TypeEnv struct { store map[string]string; outer *TypeEnv }
func NewTypeEnv() *TypeEnv { return &TypeEnv{store: make(map[string]string)} }
func (e *TypeEnv) Set(name, t string) { e.store[name] = t }
func (e *TypeEnv) Get(name string) (string, bool) {
	t, ok := e.store[name]
	if !ok && e.outer != nil { return e.outer.Get(name) }
	return t, ok
}

func GetType(expr Expression, tenv *TypeEnv) string {
	switch e := expr.(type) {
	case *BooleanLiteral: return "boolean"
	case *NewQubit: return "qubit"
	case *Identifier:
		t, ok := tenv.Get(e.Value)
		if !ok { panic("Type error: undefined variable: " + e.Value) }
		return t
	case *CallExpr:
		if e.Function == "measure" { return "boolean" }
		panic("Type error: void function cannot be used as expression")
	}
	return ""
}

type ModuleLoader func(string) (string, error)

func TypeCheck(node Node, tenv *TypeEnv, filename string, loader ModuleLoader) {
	switch n := node.(type) {
	case *Program:
		for _, stmt := range n.Stmts { TypeCheck(stmt, tenv, filename, loader) }
	case *ConstDecl:
		vt := GetType(n.Value, tenv)
		if vt != n.Type { panic(fmt.Sprintf("Type error in %s: cannot assign %s to %s", filename, vt, n.Type)) }
		tenv.Set(n.Name.Value, n.Type)
	case *ExportStmt: TypeCheck(n.Decl, tenv, filename, loader)
	case *ImportStmt:
		modSrc, err := loader(n.Module)
		if err != nil { panic("Failed to load module: " + n.Module) }
		modAst := NewParser(NewLexer(modSrc)).ParseProgram()
		modTEnv := NewTypeEnv()
		TypeCheck(modAst, modTEnv, n.Module, loader)
		for _, name := range n.Names {
			if t, ok := modTEnv.Get(name.Value); ok { tenv.Set(name.Value, t) } else {
				panic(fmt.Sprintf("Type error in %s: imported name '%s' not found", filename, name.Value))
			}
		}
	case *CallExpr:
		if n.Function == "superpose" || n.Function == "invert" || n.Function == "measure" {
			argType := GetType(n.Args[0], tenv)
			if argType != "qubit" {
				panic(fmt.Sprintf("Fatal error in %s: cannot apply quantum gate '%s' to classical type '%s'", filename, n.Function, argType))
			}
		}
	case *IfStmt:
		condType := GetType(n.Condition, tenv)
		if condType != "boolean" && condType != "qubit" { panic("Type error: if condition must be boolean or qubit") }
		TypeCheck(n.Consequence, tenv, filename, loader)
	case *ExpressionStmt:
		TypeCheck(n.Expr, tenv, filename, loader)
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
	Amps   []Complex
	NextID int
	Tally  map[string]bool
}

func NewQState() *QState {
	s := &QState{Amps: make([]Complex, 1024), Tally: make(map[string]bool)}
	s.Amps[0] = Complex{1, 0}
	return s
}

func (s *QState) Alloc() int { id := s.NextID; s.NextID++; return id }

func checkControls(i int, controls []int) bool {
	for _, c := range controls { if (i & (1 << c)) == 0 { return false } }
	return true
}

func (s *QState) applyGate(target int, controls []int, U [4]Complex) {
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

func (s *QState) H(q int, controls []int) {
	inv := 1.0 / math.Sqrt(2)
	s.applyGate(q, controls, [4]Complex{{inv, 0}, {inv, 0}, {inv, 0}, {-inv, 0}})
}

func (s *QState) X(q int, controls []int) {
	s.applyGate(q, controls, [4]Complex{{0, 0}, {1, 0}, {1, 0}, {0, 0}})
}

func (s *QState) Measure(q int, name string) bool {
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
type QubitVal struct { ID int; Name string }

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
		for _, stmt := range n.Stmts { Eval(stmt, env, qstate, controls, loader) }
	case *BlockStmt:
		for _, stmt := range n.Stmts { Eval(stmt, env, qstate, controls, loader) }
	case *ConstDecl:
		val := Eval(n.Value, env, qstate, controls, loader)
		if qval, ok := val.(*QubitVal); ok { qval.Name = n.Name.Value }
		env.Set(n.Name.Value, val)
	case *ExportStmt:
		Eval(n.Decl, env, qstate, controls, loader)
	case *IfStmt:
		cond := Eval(n.Condition, env, qstate, controls, loader)
		if qval, ok := cond.(*QubitVal); ok {
			newControls := append([]int{}, controls...)
			newControls = append(newControls, qval.ID)
			Eval(n.Consequence, env, qstate, newControls, loader)
		} else if bval, ok := cond.(*BooleanVal); ok {
			if bval.Value { Eval(n.Consequence, env, qstate, controls, loader) }
		}
	case *ExpressionStmt: Eval(n.Expr, env, qstate, controls, loader)
	case *ImportStmt:
		modSrc, _ := loader(n.Module)
		modAst := NewParser(NewLexer(modSrc)).ParseProgram()
		modEnv := NewEnv()
		Eval(modAst, modEnv, qstate, controls, loader) 
		for _, name := range n.Names {
			if val, ok := modEnv.Get(name.Value); ok { env.Set(name.Value, val) }
		}
	case *NewQubit: return &QubitVal{ID: qstate.Alloc()}
	case *BooleanLiteral: return &BooleanVal{Value: n.Value}
	case *Identifier:
		val, ok := env.Get(n.Value)
		if !ok { panic("Runtime error: undefined variable: " + n.Value) }
		return val
	case *CallExpr:
		if n.Function == "superpose" {
			q := Eval(n.Args[0], env, qstate, controls, loader).(*QubitVal)
			qstate.H(q.ID, controls)
		} else if n.Function == "invert" {
			q := Eval(n.Args[0], env, qstate, controls, loader).(*QubitVal)
			qstate.X(q.ID, controls)
		} else if n.Function == "measure" {
			q := Eval(n.Args[0], env, qstate, controls, loader).(*QubitVal)
			return &BooleanVal{Value: qstate.Measure(q.ID, q.Name)}
		}
	}
	return nil
}

// =======================================================
// 7. ENTRYPOINT (Shared by CLI and WASM)
// =======================================================

// Run evaluates the QuantumScript source code using 1000 shots
func Run(src string, filename string, loader ModuleLoader) map[string]map[bool]int {
	ast := NewParser(NewLexer(src)).ParseProgram()
	tenv := NewTypeEnv()
	TypeCheck(ast, tenv, filename, loader)

	globalTally := make(map[string]map[bool]int)
	for shot := 0; shot < 1000; shot++ {
		qstate := NewQState() 
		env := NewEnv()
		Eval(ast, env, qstate, nil, loader) 
		for name, res := range qstate.Tally {
			if globalTally[name] == nil { globalTally[name] = make(map[bool]int) }
			globalTally[name][res]++
		}
	}
	return globalTally
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
