package mcp_test

import (
	"bytes"
	"testing"

	"github.com/torresglauco/pganalytics-v3/backend/internal/mcp/transport"
)

func TestStdioTransportRead(t *testing.T) {
	input := []byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`)
	reader := bytes.NewReader(input)
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	if tr == nil {
		t.Fatal("NewStdioTransport returned nil")
	}
}

func TestStdioTransportWrite(t *testing.T) {
	input := []byte{}
	reader := bytes.NewReader(input)
	writer := &bytes.Buffer{}

	tr := transport.NewStdioTransport(reader, writer)
	data := []byte(`{"jsonrpc":"2.0","id":1,"result":{}}`)
	err := tr.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if writer.Len() == 0 {
		t.Fatal("Write did not produce output")
	}
}
