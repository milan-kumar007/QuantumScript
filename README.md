# QuantumScript

QuantumScript is a strict, natively compiled programming language that simulates wave-function collapse, quantum superposition, and entanglement right in your code. 

## 🚀 The Terminal Experience

You can install QuantumScript globally on your machine to seamlessly compile and run `.qts` files.

### Installation (Windows)
Run the PowerShell installer:
```powershell
.\install.ps1
```

### Usage
Once installed, the `qs` command is globally available!

```bash
# Scaffold a new project
qs init my-quantum-app

# Navigate to project
cd my-quantum-app

# Run the 1,000-shot hardware simulator
qs run main.qts
```

## 🌐 The Browser Experience (Quantum Websites)

You can write QuantumScript directly inside your HTML files. No build tools required! The engine compiles to WebAssembly (WASM) and runs entirely in the client's browser.

1. Drop `qs.wasm`, `wasm_exec.js`, and `qs-browser.js` into your `public/` directory.
2. Include the scripts in your `index.html`:

```html
<script src="wasm_exec.js"></script>

<script type="text/quantumscript">
    const q: qubit = new Qubit();
    superpose(q);
    measure(q);
</script>

<script src="qs-browser.js"></script>
```

When you open the webpage, the QuantumScript will execute instantly and output the wave-function collapse statistics directly to the page!

## Language Features
- **Strict Typing**: The compiler statically analyzes your code. If you try to pass a classical `boolean` into a quantum gate, the compiler will instantly reject it.
- **Entanglement (`if`)**: Placing a qubit inside an `if` statement automatically treats the block as a Quantum Controlled-NOT operation, allowing perfect entanglement.
- **Modularity**: Share precise wave-functions across multiple files using `export const q: qubit` and `import { q } from "file.qts"`.
