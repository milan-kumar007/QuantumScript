package engine

import (
	"strings"
	"testing"
)

func TestLexer(t *testing.T) {
	input := `const pi: number = 3.14; let q: qubit = new Qubit(); superpose(q);`
	l := NewLexer(input)

	expected := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{CONST, "const"},
		{IDENT, "pi"},
		{COLON, ":"},
		{NUMBER, "number"},
		{ASSIGN, "="},
		{NUMBER, "3.14"},
		{SEMI, ";"},
		{LET, "let"},
		{IDENT, "q"},
		{COLON, ":"},
		{QUBIT, "qubit"},
		{ASSIGN, "="},
		{NEW, "new"},
		{IDENT, "Qubit"},
		{LPAREN, "("},
		{RPAREN, ")"},
		{SEMI, ";"},
		{SUPERPOSE, "superpose"},
		{LPAREN, "("},
		{IDENT, "q"},
		{RPAREN, ")"},
		{SEMI, ";"},
		{EOF, ""},
	}

	for i, tt := range expected {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestParser(t *testing.T) {
	input := `
		const x: number = 5 + 5 * 10;
		if (true) {
			return x;
		}
	`
	l := NewLexer(input)
	p := NewParser(l)
	prog := p.ParseProgram()

	if prog == nil {
		t.Fatalf("ParseProgram() returned nil")
	}
	if len(prog.Stmts) != 2 {
		t.Fatalf("program.Stmts does not contain 2 statements. got=%d", len(prog.Stmts))
	}
	
	// Basic structure check
	_, ok1 := prog.Stmts[0].(*ConstDecl)
	if !ok1 {
		t.Errorf("Stmts[0] is not ConstDecl. got=%T", prog.Stmts[0])
	}
	
	_, ok2 := prog.Stmts[1].(*IfStmt)
	if !ok2 {
		t.Errorf("Stmts[1] is not IfStmt. got=%T", prog.Stmts[1])
	}
}

func TestTypeCheck(t *testing.T) {
	input := `
		const x: number = 10;
		const y: string = "hello";
		const z: number = x + 5;
	`
	l := NewLexer(input)
	p := NewParser(l)
	prog := p.ParseProgram()
	tenv := NewTypeEnv()
	
	loader := func(module string) (string, error) { return "", nil }
	
	// Should not panic
	TypeCheck(prog, tenv, "test.qts", loader)
	
	// Verify env
	if ty, ok := tenv.Get("x"); !ok || ty != "number" {
		t.Errorf("x should be number")
	}
	if ty, ok := tenv.Get("y"); !ok || ty != "string" {
		t.Errorf("y should be string")
	}
}

func TestSimulatorClassical(t *testing.T) {
	input := `
		let count: number = 0;
		for (let i = 0; i < 5; i = i + 1) {
			count = count + i;
		}
		const res: boolean = count == 10;
	`
	loader := func(module string) (string, error) { return "", nil }
	Run(input, "test.qts", loader, 0.0)
	// We just ensure it runs without panicking. The true test of output is via CLI or capturing prints,
	// but Run returns globalTally for measurements.
}

func TestSimulatorQuantum(t *testing.T) {
	input := `
		const q: qubit = new Qubit();
		superpose(q);
		measure(q);
	`
	loader := func(module string) (string, error) { return "", nil }
	res := Run(input, "test.qts", loader, 0.0)
	
	if len(res) == 0 {
		t.Fatalf("Expected measurement results, got none")
	}
	
	// For superpose(q), we expect both true and false to have roughly 500 counts each out of 1000
	counts, ok := res["q"]
	if !ok {
		t.Fatalf("Expected results for qubit 'q'")
	}
	
	if counts[true] < 100 || counts[false] < 100 {
		t.Errorf("Expected roughly 50/50 distribution, got true=%d, false=%d", counts[true], counts[false])
	}
}

func TestQASMExportLinearRouting(t *testing.T) {
	input := `
		const q0: qubit = new Qubit();
		const q1: qubit = new Qubit();
		const q2: qubit = new Qubit();
		const q3: qubit = new Qubit();
		
		if (q0) {
			X(q3);
		}
	`
	loader := func(module string) (string, error) { return "", nil }
	qasm := ExportQASM(input, "test.qts", loader, "linear")
	
	if !strings.Contains(qasm, "swap q0[0], q1[0];") {
		t.Errorf("Expected SWAP routing for linear topology, but got:\n%s", qasm)
	}
}

func TestAutomaticUncomputation(t *testing.T) {
	input := `
		function testGC() {
			let a: qubit = new Qubit();
			X(a);
		}
		testGC();
	`
	loader := func(module string) (string, error) { return "", nil }
	qasm := ExportQASM(input, "test.qts", loader, "")
	
	// The GC should inject an X gate at the end to uncompute
	// History of 'a': "x" -> adjoint is "x"
	countX := strings.Count(qasm, "x q")
	if countX != 2 {
		t.Errorf("Expected exactly 2 X gates (1 explicit, 1 from GC), found %d in:\n%s", countX, qasm)
	}
}
