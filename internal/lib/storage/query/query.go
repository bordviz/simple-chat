package query

import "strings"

func QueryToString(q string) string {
	return strings.TrimSpace(
		strings.ReplaceAll(strings.ReplaceAll(q, "\t", ""), "\n", " "),
	)
}
