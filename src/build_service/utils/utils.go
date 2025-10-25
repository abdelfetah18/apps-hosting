package utils

import (
	"strings"
)

func ToImageName(appName string) string {
	return strings.ToLower(strings.ReplaceAll(appName, " ", "-"))
}

func ToK8sLabelValue(appName string) string {
	return strings.ToLower(strings.ReplaceAll(appName, " ", "-"))
}

func ToK8sJobName(appName string) string {
	return ToK8sLabelValue(appName) + "-job"
}
