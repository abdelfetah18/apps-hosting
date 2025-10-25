package kubernetes

import (
	"deploy/utils"

	v1Apps "k8s.io/api/apps/v1"
	v1Core "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/intstr"
)

func GenerateDeploymentObject(appId, appName, imageURL string, envVars []v1Core.EnvVar) v1Apps.Deployment {
	return v1Apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: utils.ToK8sDeploymentName(appName),
			Labels: map[string]string{
				"app_name": utils.ToK8sLabelValue(appName),
				"app_id":   appId,
			},
		},
		Spec: v1Apps.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app_name": utils.ToK8sLabelValue(appName),
					"app_id":   appId,
				},
			},
			Template: v1Core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app_name": utils.ToK8sLabelValue(appName),
						"app_id":   appId,
					},
				},
				Spec: v1Core.PodSpec{
					Containers: []v1Core.Container{
						{
							Name:  utils.ToK8sContainerName(appName),
							Image: imageURL,
							Ports: []v1Core.ContainerPort{
								{ContainerPort: 3000},
							},
							Env: envVars,
						},
					},
				},
			},
		},
	}

}

func GenerateServiceObject(namespace, appId, appName string) v1Core.Service {
	return v1Core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      utils.ToK8sServiceName(appName),
			Namespace: namespace,
			Labels: map[string]string{
				"app_name": utils.ToK8sLabelValue(appName),
				"app_id":   appId,
			},
		},
		Spec: v1Core.ServiceSpec{
			Selector: map[string]string{
				"app_name": utils.ToK8sLabelValue(appName),
				"app_id":   appId,
			},
			Ports: []v1Core.ServicePort{
				{
					Port:       80,
					TargetPort: intstr.FromInt(3000),
					Protocol:   v1Core.ProtocolTCP,
				},
			},
		},
	}
}

func GenerateIngressObject(namespace, appId, appName, host, serviceName string) networkingv1.Ingress {
	pathType := networkingv1.PathTypePrefix

	ingressClassName := "nginx"

	return networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: utils.ToK8sIngressName(appName),
			Labels: map[string]string{
				"app_name": utils.ToK8sLabelValue(appName),
				"app_id":   appId,
			},
			Namespace: namespace,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/force-ssl-redirect": "true",
				"kubernetes.io/ingress.global-static-ip-name":    "gke-ingress-ip",
			},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ingressClassName,
			TLS: []networkingv1.IngressTLS{
				{
					Hosts:      []string{host},
					SecretName: "tls-wildcard.apps-hosting.com",
				},
			},
			Rules: []networkingv1.IngressRule{
				{
					Host: host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: serviceName,
											Port: networkingv1.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
