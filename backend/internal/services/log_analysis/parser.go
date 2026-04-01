package log_analysis

import (
	"regexp"
	"strings"
)

type LogParser struct {
	patterns map[LogCategory]*regexp.Regexp
}

func NewLogParser() *LogParser {
	parser := &LogParser{
		patterns: make(map[LogCategory]*regexp.Regexp),
	}

	for category, patternList := range LogCategoryPatterns {
		combined := "(" + strings.Join(patternList, "|") + ")"
		parser.patterns[category] = regexp.MustCompile(combined)
	}

	return parser
}

func (lp *LogParser) ClassifyLog(message string) LogCategory {
	for category, pattern := range lp.patterns {
		if pattern.MatchString(message) {
			return category
		}
	}

	if strings.Contains(strings.ToLower(message), "error") {
		return CategoryDatabaseError
	}

	if strings.Contains(strings.ToLower(message), "warning") {
		return CategoryWarning
	}

	return CategoryInfo
}

func (lp *LogParser) ExtractMetadata(message string) map[string]interface{} {
	metadata := make(map[string]interface{})

	durationRegex := regexp.MustCompile(`duration: ([\d.]+) ms`)
	if matches := durationRegex.FindStringSubmatch(message); len(matches) > 1 {
		metadata["duration"] = matches[1]
	}

	tableRegex := regexp.MustCompile(`relation "([^"]+)"`)
	if matches := tableRegex.FindStringSubmatch(message); len(matches) > 1 {
		metadata["table"] = matches[1]
	}

	return metadata
}
