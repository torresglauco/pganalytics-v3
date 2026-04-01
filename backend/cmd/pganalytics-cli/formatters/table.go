package formatters

import (
	"fmt"
	"strings"
)

type TableFormatter struct {
	headers []string
	rows    [][]string
}

func (f *TableFormatter) AddHeader(headers ...string) {
	f.headers = headers
}

func (f *TableFormatter) AddRow(values ...string) {
	f.rows = append(f.rows, values)
}

func (f *TableFormatter) Format(data interface{}) string {
	// Calculate column widths
	colWidths := make([]int, len(f.headers))

	for i, header := range f.headers {
		colWidths[i] = len(header)
	}

	for _, row := range f.rows {
		for i, val := range row {
			if len(val) > colWidths[i] {
				colWidths[i] = len(val)
			}
		}
	}

	var output strings.Builder

	// Write headers
	for i, header := range f.headers {
		output.WriteString(fmt.Sprintf("%-*s", colWidths[i], header))
		if i < len(f.headers)-1 {
			output.WriteString(" | ")
		}
	}
	output.WriteString("\n")

	// Write separator
	for i, width := range colWidths {
		output.WriteString(strings.Repeat("-", width))
		if i < len(colWidths)-1 {
			output.WriteString("-+-")
		}
	}
	output.WriteString("\n")

	// Write rows
	for _, row := range f.rows {
		for i, val := range row {
			output.WriteString(fmt.Sprintf("%-*s", colWidths[i], val))
			if i < len(row)-1 {
				output.WriteString(" | ")
			}
		}
		output.WriteString("\n")
	}

	return output.String()
}
