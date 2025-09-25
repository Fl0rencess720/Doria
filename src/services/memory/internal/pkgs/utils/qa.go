package utils

import "strings"

func BuildQAPair(question, answer string) string {
	var builder strings.Builder

	builder.WriteString(question)
	builder.WriteString("\n\n\n\n")
	builder.WriteString(answer)

	return builder.String()
}
