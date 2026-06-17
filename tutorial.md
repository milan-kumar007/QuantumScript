# QuantumScript: The Official Tutorial

Welcome to **QuantumScript (.qts)**—the TypeScript of the Quantum Web. 
QuantumScript is a lightweight, natively compiled language designed to bridge the gap between classical web development and quantum computing. It runs natively in your terminal via Go, and natively in your browser via WebAssembly (`qs.wasm`).

This tutorial will take you from zero to building your first quantum web algorithms.

---

## 1. The Basics
QuantumScript is strictly typed, just like TypeScript. You declare variables using `const` or `let`.

### Classical Types
You have access to standard classical logic to build your algorithms:
```javascript
const pi: number = 3.14159;
let counter: number = 0;
const isReady: boolean = true;
```
*Note: `number` supports all basic mathematical operators (`+`, `-`, `*`, `/`).*

### The Qubit
The core of the language is the `qubit` type. When you allocate a new qubit, it starts in the absolute `|0⟩` state.
```javascript
const q1: qubit = new Qubit();
```

---

## 2. Universal Quantum Gates
QuantumScript comes with a complete embedded linear algebra simulator. You can manipulate qubits using standard quantum gates:

### The Core Gates
* **`superpose(q)`**: Also known as the Hadamard (H) gate. Puts a qubit into a perfect 50/50 superposition of `|0⟩` and `|1⟩`.
* **`invert(q)`**: Also known as the Pauli-X gate. Flips a qubit from `|0⟩` to `|1⟩`.
* **`measure(q)`**: Collapses the quantum wave-function into a classical `boolean` (`true` for `|1⟩`, `false` for `|0⟩`).

### The Advanced Standard Library
* **Pauli Gates**: `Y(q)`, `Z(q)`
* **Phase Gates**: `S(q)`, `T(q)`
* **Rotation Gates**: `Rx(theta, q)`, `Ry(theta, q)`, `Rz(theta, q)`

*Example: Rotating a qubit by Pi*
```javascript
const q: qubit = new Qubit();
const angle: number = 3.14159;
Rx(angle, q); // Rotates the qubit 180 degrees around the X-axis
```

---

## 3. Arrays and Loops
Building scalable quantum algorithms (like Shor's or Grover's) requires manipulating many qubits at once.

### Initializing Quantum Registers
You can allocate an array of qubits using standard syntax:
```javascript
const qReg: qubit[] = new Qubit[5]; // Creates 5 qubits
```

### Classical Control Flow
QuantumScript fully supports `for` and `while` loops for classical iteration:
```javascript
// Put all 5 qubits into superposition
for (let i = 0; i < 5; i = i + 1) {
    superpose(qReg[i]);
}
```

---

## 4. Custom Functions
Make your quantum algorithms modular and reusable by declaring functions.
```javascript
function flipAndMeasure(q: qubit): boolean {
    invert(q);
    return measure(q);
}

const myQubit: qubit = new Qubit();
const result: boolean = flipAndMeasure(myQubit);
```

---

## 5. Entanglement (The CNOT Gate)
Quantum entanglement is handled elegantly through standard `if` statements. 
When you pass a `qubit` into an `if` condition, the compiler interprets the entire block as a **Quantum Controlled Operation**.

To create a standard CNOT (Controlled-NOT) gate:
```javascript
const control: qubit = new Qubit();
const target: qubit = new Qubit();

superpose(control); // Put control into superposition

// This acts as a CNOT gate!
if (control) {
    invert(target);
}
```
*If `control` is measured as `|1⟩`, the `target` is inverted. Because `control` is in superposition, both qubits become mathematically entangled!*

---

## 6. Running Your Code

### In the Terminal
You can run any `.qts` file locally using the CLI engine. It will simulate the circuit 1000 times and print the statistical probability of your wave-functions!
```bash
qs run my_algorithm.qts
```

### In the Browser (The Quantum Web)
Because QuantumScript compiles to WebAssembly, you can run `.qts` scripts directly inside your HTML!
```html
<script src="wasm_exec.js"></script>
<script src="qs-browser.js"></script>
<script>
    // Execute raw QuantumScript in the browser!
    const code = `
        const q: qubit = new Qubit();
        superpose(q);
        measure(q);
    `;
    const results = window.evaluateQuantumScript(code);
    console.log(results);
</script>
```

---
**Happy Quantum Coding!** Welcome to the future of the web.
