package formatters

import (
	"encoding/json"
	"fmt"
)

type JSONFormatter struct{}

func (f *JSONFormatter) Format(data interface{}) string {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return string(b)
}
