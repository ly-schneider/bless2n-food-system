package utils

import "strings"

func IsMobile(ua *string) bool {
	return strings.Contains(*ua, "Android")
}
