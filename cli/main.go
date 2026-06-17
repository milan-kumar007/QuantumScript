package main

import (
	"fmt"
	"os"
	"path/filepath"
	"quantumscript/engine"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "version":
		fmt.Println("QuantumScript Compiler v1.0.0")
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
			fmt.Println("Usage: qs run <file.qts>")
			return
		}
		filename := os.Args[2]
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
			
			results := engine.Run(string(src), filename, loader)

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
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("QuantumScript CLI")
	fmt.Println("Usage:")
	fmt.Println("  qs init <project>   Initialize a new project")
	fmt.Println("  qs run <file.qts>   Execute a QuantumScript file")
	fmt.Println("  qs version          Show compiler version")
}
