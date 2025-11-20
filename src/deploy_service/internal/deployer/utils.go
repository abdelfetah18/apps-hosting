package deployer

import "strings"

func ToK8sLabelValue(appName string) string {
	return strings.ToLower(strings.ReplaceAll(appName, " ", "-"))
}

func ToK8sDeploymentName(appName string) string {
	return ToK8sLabelValue(appName) + "-deployment"
}

func ToK8sContainerName(appName string) string {
	return ToK8sLabelValue(appName) + "-container"
}

func ToK8sServiceName(appName string) string {
	return ToK8sLabelValue(appName) + "-service"
}

func ToK8sIngressName(appName string) string {
	return ToK8sLabelValue(appName) + "-ingress"
}
