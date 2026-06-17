package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"quantumscript/engine"
)

type Request struct {
	RPC    string          `json:"jsonrpc"`
	ID     *int            `json:"id,omitempty"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type Response struct {
	RPC    string      `json:"jsonrpc"`
	ID     *int        `json:"id,omitempty"`
	Result interface{} `json:"result,omitempty"`
	Error  interface{} `json:"error,omitempty"`
}

type Notification struct {
	RPC    string      `json:"jsonrpc"`
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

var errorRegex = regexp.MustCompile(`error at line (\d+), col (\d+): (.+)`)

func StartServer() {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Read headers
		var contentLength int
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					return // Client closed connection
				}
				continue
			}
			line = strings.TrimSpace(line)
			if line == "" {
				break
			}
			if strings.HasPrefix(line, "Content-Length:") {
				fmt.Sscanf(line, "Content-Length: %d", &contentLength)
			}
		}

		if contentLength == 0 {
			continue
		}

		// Read payload
		payload := make([]byte, contentLength)
		_, err := io.ReadFull(reader, payload)
		if err != nil {
			continue
		}

		var req Request
		if err := json.Unmarshal(payload, &req); err != nil {
			continue
		}

		handleRequest(req)
	}
}

func handleRequest(req Request) {
	switch req.Method {
	case "initialize":
		sendResponse(req.ID, map[string]interface{}{
			"capabilities": map[string]interface{}{
				"textDocumentSync": 1, // Full sync
			},
		})
	case "textDocument/didOpen":
		var params struct {
			TextDocument struct {
				URI  string `json:"uri"`
				Text string `json:"text"`
			} `json:"textDocument"`
		}
		json.Unmarshal(req.Params, &params)
		validateDocument(params.TextDocument.URI, params.TextDocument.Text)
	case "textDocument/didChange":
		var params struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
			ContentChanges []struct {
				Text string `json:"text"`
			} `json:"contentChanges"`
		}
		json.Unmarshal(req.Params, &params)
		if len(params.ContentChanges) > 0 {
			validateDocument(params.TextDocument.URI, params.ContentChanges[0].Text)
		}
	}
}

func validateDocument(uri, text string) {
	var diagnostics []map[string]interface{}

	func() {
		defer func() {
			if r := recover(); r != nil {
				errStr := fmt.Sprintf("%v", r)
				
				// Try to extract line and column
				matches := errorRegex.FindStringSubmatch(errStr)
				var line, col int
				var msg string
				if len(matches) == 4 {
					line, _ = strconv.Atoi(matches[1])
					col, _ = strconv.Atoi(matches[2])
					msg = matches[3]
					line = line - 1 // LSP is 0-indexed
				} else {
					line = 0
					col = 0
					msg = errStr
				}

				diagnostics = append(diagnostics, map[string]interface{}{
					"range": map[string]interface{}{
						"start": map[string]int{"line": line, "character": col},
						"end":   map[string]int{"line": line, "character": col + 1},
					},
					"severity": 1, // Error
					"source":   "quantumscript",
					"message":  msg,
				})
			}
		}()

		// Fake module loader for LSP
		loader := func(module string) (string, error) {
			return "", fmt.Errorf("LSP cannot resolve imports yet")
		}

		ast := engine.NewParser(engine.NewLexer(text)).ParseProgram()
		tenv := engine.NewTypeEnv()
		engine.TypeCheck(ast, tenv, "lsp", loader)
	}()

	sendNotification("textDocument/publishDiagnostics", map[string]interface{}{
		"uri":         uri,
		"diagnostics": diagnostics,
	})
}

func sendResponse(id *int, result interface{}) {
	res := Response{
		RPC:    "2.0",
		ID:     id,
		Result: result,
	}
	send(res)
}

func sendNotification(method string, params interface{}) {
	notif := Notification{
		RPC:    "2.0",
		Method: method,
		Params: params,
	}
	send(notif)
}

func send(msg interface{}) {
	data, _ := json.Marshal(msg)
	fmt.Printf("Content-Length: %d\r\n\r\n%s", len(data), string(data))
}
