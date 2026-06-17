const { LanguageClient, TransportKind } = require('vscode-languageclient/node');

let client;

function activate(context) {
    let serverExecutable = {
        command: 'qs',
        args: ['lsp'],
        options: {
            env: process.env // Inherit the environment to find qs on the PATH
        }
    };

    let serverOptions = {
        run: serverExecutable,
        debug: serverExecutable
    };

    let clientOptions = {
        documentSelector: [{ scheme: 'file', language: 'quantumscript' }],
    };

    client = new LanguageClient(
        'quantumscript-lsp',
        'QuantumScript Language Server',
        serverOptions,
        clientOptions
    );

    client.start();
}

function deactivate() {
    if (!client) {
        return undefined;
    }
    return client.stop();
}

module.exports = {
    activate,
    deactivate
};
