package fanstr

import "testing"

func TestFormatFieldName(t *testing.T) {
	tests := []struct {
		input    string
		args     []string
		expected string
	}{
		{
			input:    "Hello {name|Guest}!",
			args:     []string{"name", "Tom"},
			expected: "Hello Tom!",
		},
		{
			input:    "Hello {name|Guest}!",
			args:     []string{}, // 没有name
			expected: "Hello Guest!",
		},
		{
			input:    "Missing {field|DefaultValue}",
			args:     []string{"other", "X"},
			expected: "Missing DefaultValue",
		},
		{
			input:    "Curly braces: {{ and }}",
			args:     []string{},
			expected: "Curly braces: { and }",
		},
		{
			input:    "No Default {field}",
			args:     []string{},
			expected: "No Default {field}",
		},
	}

	for _, test := range tests {
		actual := FormatFieldName(test.input, test.args...)
		if actual != test.expected {
			t.Errorf("Input: %q\nArgs: %v\nExpected: %q\nActual:   %q", test.input, test.args, test.expected, actual)
		}
	}
}
