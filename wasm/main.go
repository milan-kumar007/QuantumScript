package main

import (
	"fmt"
	"syscall/js"
	"quantumscript/engine"
)

func evaluateQuantumScript(this js.Value, p []js.Value) interface{} {
	if len(p) == 0 {
		return "Error: no source code provided"
	}
	src := p[0].String()
	
	noiseLevel := 0.0
	if len(p) > 1 && p[1].Type() == js.TypeNumber {
		noiseLevel = p[1].Float()
	}
	loader := func(moduleName string) (string, error) {
		return "", fmt.Errorf("module loading not fully supported in simple WASM demo: %s", moduleName)
	}

	var output string
	func() {
		defer func() {
			if r := recover(); r != nil {
				output = fmt.Sprintf("Error: %v", r)
			}
		}()
		
		results := engine.Run(src, "browser.qs", loader, noiseLevel)

		output += "QuantumScript Simulator Results (1000 Shots):\n"
		for name, counts := range results {
			total := counts[true] + counts[false]
			if total == 0 {
				continue
			}
			p1 := float64(counts[true]) / float64(total) * 100
			p0 := float64(counts[false]) / float64(total) * 100
			output += fmt.Sprintf("Qubit [%s] -> |1⟩: %.1f%% |0⟩: %.1f%%\n", name, p1, p0)
		}
	}()
	return output
}

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("evaluateQuantumScript", js.FuncOf(evaluateQuantumScript))
	<-c
}
