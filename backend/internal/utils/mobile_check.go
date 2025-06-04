package utils

import "strings"

func IsMobile(ua string) bool { // very naive; TODO: improve this once the native app is ready
	return strings.Contains(ua, "Android") || strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad")
}
