package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"quantumscript/engine"
	"quantumscript/lsp"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "version":
		fmt.Println("QuantumScript Compiler v4.0.0")
	case "init":
		if len(os.Args) < 3 {
			fmt.Println("Usage: qs init <project-name>")
			return
		}
		projectName := os.Args[2]
		os.MkdirAll(projectName, 0755)

		mainCode := `import { sharedQubit } from "module.qts";

const localQubit: qubit = new Qubit();

if (sharedQubit) {
    invert(localQubit);
}

measure(sharedQubit);
measure(localQubit);
`
		moduleCode := `export const sharedQubit: qubit = new Qubit();
superpose(sharedQubit);
`
		os.WriteFile(filepath.Join(projectName, "main.qts"), []byte(mainCode), 0644)
		os.WriteFile(filepath.Join(projectName, "module.qts"), []byte(moduleCode), 0644)
		fmt.Printf("Initialized QuantumScript project '%s'.\n", projectName)
		fmt.Printf("Run 'cd %s' and 'qs run main.qts' to see it in action.\n", projectName)
	case "run":
		if len(os.Args) < 3 {
			fmt.Println("Usage: qs run [--noise=0.01] <file.qts>")
			return
		}
		
		noiseLevel := 0.0
		var filename string
		for _, arg := range os.Args[2:] {
			if strings.HasPrefix(arg, "--noise=") {
				noiseStr := strings.TrimPrefix(arg, "--noise=")
				fmt.Sscanf(noiseStr, "%f", &noiseLevel)
			} else {
				filename = arg
			}
		}
		
		if filename == "" {
			fmt.Println("Usage: qs run [--noise=0.01] <file.qts>")
			return
		}
		
		src, err := os.ReadFile(filename)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", filename, err)
			os.Exit(1)
		}

		loader := func(moduleName string) (string, error) {
			dir := filepath.Dir(filename)
			modPath := filepath.Join(dir, moduleName)
			b, err := os.ReadFile(modPath)
			return string(b), err
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println(r)
					os.Exit(1)
				}
			}()
			
			results := engine.Run(string(src), filename, loader, noiseLevel)

			fmt.Println("QuantumScript Simulator Results (1000 Shots):")
			for name, counts := range results {
				total := counts[true] + counts[false]
				if total == 0 {
					continue
				}
				p1 := float64(counts[true]) / float64(total) * 100
				p0 := float64(counts[false]) / float64(total) * 100
				fmt.Printf("Qubit [%s] -> |1⟩: %.1f%% |0⟩: %.1f%%\n", name, p1, p0)
			}
		}()
	case "export-qasm":
		if len(os.Args) < 3 {
			fmt.Println("Usage: qs export-qasm [--topology=linear] <file.qts>")
			return
		}
		
		topology := ""
		var filename string
		for _, arg := range os.Args[2:] {
			if strings.HasPrefix(arg, "--topology=") {
				topology = strings.TrimPrefix(arg, "--topology=")
			} else {
				filename = arg
			}
		}
		
		if filename == "" {
			fmt.Println("Usage: qs export-qasm [--topology=linear] <file.qts>")
			return
		}
		
		src, err := os.ReadFile(filename)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", filename, err)
			os.Exit(1)
		}

		loader := func(moduleName string) (string, error) {
			dir := filepath.Dir(filename)
			modPath := filepath.Join(dir, moduleName)
			b, err := os.ReadFile(modPath)
			return string(b), err
		}

		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Println(r)
					os.Exit(1)
				}
			}()
			
			qasm := engine.ExportQASM(string(src), filename, loader, topology)
			
			outFilename := filename + ".qasm"
			os.WriteFile(outFilename, []byte(qasm), 0644)
			fmt.Printf("Successfully exported OpenQASM to %s\n", outFilename)
			fmt.Println("\n" + qasm)
		}()
	case "lsp":
		lsp.StartServer()
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("QuantumScript CLI")
	fmt.Println("Usage:")
	fmt.Println("  qs init <project>   Initialize a new project")
	fmt.Println("  qs run [--noise=0.01] <file.qts>         Execute a QuantumScript file")
	fmt.Println("  qs export-qasm [--topology=linear] <file.qts> Export to OpenQASM 3.0")
	fmt.Println("  qs lsp                    Start the Language Server")
	fmt.Println("  qs version                Show compiler version")
}
