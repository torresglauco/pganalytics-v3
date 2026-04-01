package formatters

import (
	"encoding/csv"
	"fmt"
	"strings"
)

type CSVFormatter struct {
	headers []string
	rows    [][]string
}

func (f *CSVFormatter) AddHeader(headers ...string) {
	f.headers = headers
}

func (f *CSVFormatter) AddRow(values ...string) {
	f.rows = append(f.rows, values)
}

func (f *CSVFormatter) Format(data interface{}) string {
	var output strings.Builder
	writer := csv.NewWriter(&output)

	writer.Write(f.headers)
	writer.WriteAll(f.rows)
	writer.Flush()

	if err := writer.Error(); err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	return output.String()
}
