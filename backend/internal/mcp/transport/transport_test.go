package transport

import (
	"bytes"
	"testing"
)

// TestStdioTransportRead tests reading from transport
func TestStdioTransportRead(t *testing.T) {
	data := []byte("test data\n")
	reader := bytes.NewReader(data)
	writer := &bytes.Buffer{}

	tr := NewStdioTransport(reader, writer)
	result, err := tr.Read()

	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Read returned empty data")
	}

	if !bytes.Contains(result, []byte("test data")) {
		t.Errorf("Read data mismatch: got %s", string(result))
	}
}

// TestStdioTransportWrite tests writing to transport
func TestStdioTransportWrite(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	writer := &bytes.Buffer{}

	tr := NewStdioTransport(reader, writer)
	data := []byte("test output")
	err := tr.Write(data)

	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	output := writer.String()
	if !bytes.Contains([]byte(output), data) {
		t.Errorf("Write data mismatch: got %s", output)
	}

	// Verify newline was added
	if !bytes.HasSuffix([]byte(output), []byte("\n")) {
		t.Error("Write did not add newline")
	}
}

// TestStdioTransportGetReader tests getting the underlying reader
func TestStdioTransportGetReader(t *testing.T) {
	data := []byte("test data\n")
	reader := bytes.NewReader(data)
	writer := &bytes.Buffer{}

	tr := NewStdioTransport(reader, writer)
	bufReader := tr.GetReader()

	if bufReader == nil {
		t.Fatal("GetReader returned nil")
	}

	// Verify we can read from the returned reader
	result, err := bufReader.ReadBytes('\n')
	if err != nil {
		t.Fatalf("Failed to read from returned reader: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("ReadBytes returned empty data")
	}
}

// TestStdioTransportClose tests closing the transport
func TestStdioTransportClose(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	writer := &bytes.Buffer{}

	tr := NewStdioTransport(reader, writer)
	err := tr.Close()

	// Close should not error even though it's a no-op
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

// TestStdioTransportMultipleWrites tests multiple sequential writes
func TestStdioTransportMultipleWrites(t *testing.T) {
	reader := bytes.NewReader([]byte{})
	writer := &bytes.Buffer{}

	tr := NewStdioTransport(reader, writer)

	data1 := []byte("first")
	data2 := []byte("second")
	data3 := []byte("third")

	if err := tr.Write(data1); err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	if err := tr.Write(data2); err != nil {
		t.Fatalf("Second write failed: %v", err)
	}

	if err := tr.Write(data3); err != nil {
		t.Fatalf("Third write failed: %v", err)
	}

	output := writer.String()
	if !bytes.Contains([]byte(output), data1) {
		t.Error("First write not in output")
	}

	if !bytes.Contains([]byte(output), data2) {
		t.Error("Second write not in output")
	}

	if !bytes.Contains([]byte(output), data3) {
		t.Error("Third write not in output")
	}
}

// TestStdioTransportJSONRPCFormat tests proper JSON-RPC format handling
func TestStdioTransportJSONRPCFormat(t *testing.T) {
	jsonData := []byte(`{"jsonrpc":"2.0","id":1,"method":"test","params":{}}` + "\n")
	reader := bytes.NewReader(jsonData)
	writer := &bytes.Buffer{}

	tr := NewStdioTransport(reader, writer)
	result, err := tr.Read()

	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if !bytes.Contains(result, []byte("jsonrpc")) {
		t.Fatal("JSON-RPC content not preserved")
	}
}
