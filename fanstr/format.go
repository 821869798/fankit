package fanstr

import (
	"strings"
)

// FormatFieldName FormatFieldName("File {file} had error {error}", "file", file, "error", err)
func FormatFieldName(format string, args ...string) []string {
	fields := strings.Fields(format)
	formatMapping := make(map[string]string)
	for i := 0; i < len(args); i += 2 {
		key := "{" + args[i] + "}"
		value := args[i+1]
		formatMapping[key] = value
	}
	for i, v := range fields {
		if value, ok := formatMapping[v]; ok {
			fields[i] = value
		}
	}

	return fields
}
