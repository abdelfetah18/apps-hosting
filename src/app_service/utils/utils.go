package utils

import (
	"strings"
)

func GetDomainName(appName string) string {
	return strings.ToLower(strings.ReplaceAll(appName, " ", "-")) + ".apps-hosting.com"
}
