// qs-browser.js
(async function() {
    console.log("Initializing QuantumScript Browser Runtime...");

    const go = new Go(); // Provided by wasm_exec.js
    const result = await WebAssembly.instantiateStreaming(fetch("qs.wasm"), go.importObject);
    go.run(result.instance); // Starts the Go WASM runtime

    // Find all <script type="text/quantumscript">
    const scripts = document.querySelectorAll('script[type="text/quantumscript"]');
    
    scripts.forEach((script) => {
        let code = script.textContent;
        if (script.src) {
            fetch(script.src)
                .then(r => r.text())
                .then(code => execute(code));
        } else {
            execute(code);
        }
    });

    function execute(code) {
        if (!code || code.trim() === "") return;
        console.log("Executing QuantumScript:\n", code);
        
        // evaluateQuantumScript is exposed by wasm/main.go
        const output = window.evaluateQuantumScript(code);
        console.log(output);
        
        // Display on the page
        const pre = document.createElement("pre");
        pre.style.background = "#1e1e1e";
        pre.style.color = "#00ffcc";
        pre.style.padding = "15px";
        pre.style.borderRadius = "8px";
        pre.style.border = "1px solid #333";
        pre.style.fontFamily = "monospace";
        pre.innerText = output;
        document.body.appendChild(pre);
    }
})();
