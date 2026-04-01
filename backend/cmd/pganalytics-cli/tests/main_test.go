package tests

import (
	"bytes"
	"testing"
	"github.com/spf13/cobra"
)

func TestVersionFlag(t *testing.T) {
	cmd := &cobra.Command{
		Use:     "pganalytics",
		Version: "0.1.0",
		Run: func(cmd *cobra.Command, args []string) {
			// Empty run
		},
	}

	cmd.SetArgs([]string{"--version"})
	var out bytes.Buffer
	cmd.SetOut(&out)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !bytes.Contains(out.Bytes(), []byte("0.1.0")) {
		t.Errorf("Expected version in output, got: %s", out.String())
	}
}
