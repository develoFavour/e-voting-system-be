package services

import "strings"

// NormalizeMatricNumber trims surrounding whitespace and stores matric numbers
// in a consistent uppercase form.
func NormalizeMatricNumber(matricNumber string) string {
	return strings.ToUpper(strings.TrimSpace(matricNumber))
}
