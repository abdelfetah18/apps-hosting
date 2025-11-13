package utils

import (
	"strings"
)

func GetDomainName(appName string) string {
	return strings.ToLower(strings.ReplaceAll(appName, " ", "-")) + ".apps-hosting.com"
}

func SafeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
