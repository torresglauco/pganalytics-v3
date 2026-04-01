package transport

import (
	"bufio"
	"io"
)

type StdioTransport struct {
	reader *bufio.Reader
	writer io.Writer
}

func NewStdioTransport(r io.Reader, w io.Writer) *StdioTransport {
	return &StdioTransport{
		reader: bufio.NewReader(r),
		writer: w,
	}
}

func (t *StdioTransport) Read() ([]byte, error) {
	line, err := t.reader.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return line, nil
}

func (t *StdioTransport) Write(data []byte) error {
	_, err := t.writer.Write(data)
	if err != nil {
		return err
	}
	_, err = t.writer.Write([]byte("\n"))
	return err
}

func (t *StdioTransport) Close() error {
	return nil
}

// GetReader returns the underlying buffered reader
func (t *StdioTransport) GetReader() *bufio.Reader {
	return t.reader
}
