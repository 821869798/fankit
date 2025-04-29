package fanstr

import (
	"strings"
)

// FormatFieldNameMap replaces {key} style placeholders using a provided map.
func FormatFieldNameMap(input string, vars map[string]string) string {
	var builder strings.Builder
	runes := []rune(input)
	n := len(runes)

	for i := 0; i < n; {
		if runes[i] == '{' {
			if i+1 < n && runes[i+1] == '{' {
				builder.WriteRune('{')
				i += 2
			} else {
				j := i + 1
				for j < n && runes[j] != '}' {
					j++
				}
				if j < n {
					placeholder := string(runes[i+1 : j])
					key, def := splitKeyDefault(placeholder)
					if val, ok := vars[key]; ok {
						builder.WriteString(val)
					} else {
						builder.WriteString(def)
					}
					i = j + 1
				} else {
					builder.WriteRune(runes[i])
					i++
				}
			}
		} else if runes[i] == '}' {
			if i+1 < n && runes[i+1] == '}' {
				builder.WriteRune('}')
				i += 2
			} else {
				builder.WriteRune(runes[i])
				i++
			}
		} else {
			builder.WriteRune(runes[i])
			i++
		}
	}

	return builder.String()
}

// FormatFieldName replaces {key} style placeholders in input with key-value string pairs.
// Supports {key|default} to specify fallback default values.
func FormatFieldName(input string, args ...string) string {
	vars := make(map[string]string)

	argLen := len(args)
	for i := 0; i < argLen-1; i += 2 {
		key := args[i]
		val := args[i+1]
		vars[key] = val
	}

	return FormatFieldNameMap(input, vars)
}

// 拆分 {name|default} 结构
func splitKeyDefault(input string) (key string, def string) {
	parts := strings.SplitN(input, "|", 2)
	key = parts[0]
	if len(parts) == 2 {
		def = parts[1]
	} else {
		def = "{" + key + "}" // 保持原样
	}
	return
}
