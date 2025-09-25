package utils

import "strings"

func BuildQAPair(question, answer string) string {
	var builder strings.Builder

	builder.WriteString("Question: ")
	builder.WriteString(question)
	builder.WriteString("\n")
	builder.WriteString("Answer: ")
	builder.WriteString(answer)
	builder.WriteString("\n\n")

	return builder.String()
}
