import { sharedQubit } from "module.qs";

const localQubit: qubit = new Qubit();

// Entangle localQubit with sharedQubit using a Quantum CNOT logic
if (sharedQubit) {
    invert(localQubit);
}

measure(sharedQubit);
measure(localQubit);