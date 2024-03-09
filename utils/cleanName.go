package utils

import "strings"

func CleanName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
