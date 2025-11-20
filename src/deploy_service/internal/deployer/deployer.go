package deployer

import (
	"context"
	"fmt"
	"os"

	"apps-hosting.com/logging"

	v1Apps "k8s.io/api/apps/v1"
	v1Core "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

// FIXME: must be a dynamic value
const NAMESPACE = "default"

type Deployer struct {
	kubernetesClient *kubernetes.Clientset
	logger           logging.ServiceLogger
}

func NewDeployer(kubernetesClient *kubernetes.Clientset) Deployer {
	return Deployer{kubernetesClient: kubernetesClient}
}

func (d *Deployer) Deploy(appId, appName, domainName, imageUrl string, envVars []v1Core.EnvVar) error {
	labels := map[string]string{
		"app_name": ToK8sLabelValue(appName),
		"app_id":   appId,
	}

	// 1. create deployment resource
	err := d.deployImage(appName, imageUrl, labels, envVars)
	if err != nil {
		return err
	}

	// 2. expose the app to the cluster network
	serviceName, err := d.exposeAppInternally(appName, labels)
	if err != nil {
		return err
	}

	// 3. expsoing http/https routes from outside cluster to cluster network
	err = d.exposeAppExternally(appName, domainName, *serviceName, labels)
	if err != nil {
		return err
	}

	return nil
}

func (d *Deployer) Destroy(appName string) error {
	err := d.unDeployImage(appName)
	if err != nil {
		return err
	}

	err = d.unExposeAppInternally(appName)
	if err != nil {
		return err
	}

	err = d.unExposeAppExternally(appName)
	if err != nil {
		return err
	}

	return nil
}

func (d *Deployer) deployImage(appName, imageURL string, labels map[string]string, envVars []v1Core.EnvVar) error {
	deploymentObject := d.generateDeploymentObject(appName, imageURL, labels, envVars)
	deploymentsClient := d.kubernetesClient.AppsV1().Deployments(NAMESPACE)
	_, err := deploymentsClient.Create(context.Background(), &deploymentObject, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to deploy the image: %v", err)
	}

	d.logger.LogInfoF("Deployment created successfully in namespace %s with image: %s", NAMESPACE, imageURL)
	return nil
}

func (d *Deployer) exposeAppInternally(appName string, labels map[string]string) (*string, error) {
	d.logger.LogInfo("Generating kubernetes service object...")
	serviceObject := d.generateServiceObject(NAMESPACE, appName, labels)
	_, err := d.kubernetesClient.
		CoreV1().
		Services(NAMESPACE).
		Create(context.TODO(), &serviceObject, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return &serviceObject.Name, nil
}

func (d *Deployer) exposeAppExternally(appName, domainName, serviceName string, labels map[string]string) error {
	d.logger.LogInfo("Generating kubernetes ingress object...")

	ingressObject := d.generateIngressObject(appName, domainName, serviceName, labels)
	_, err := d.kubernetesClient.NetworkingV1().Ingresses(NAMESPACE).Create(context.TODO(), &ingressObject, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create ingress: %w", err)
	}

	d.logger.LogInfoF("Ingress %q created in namespace %q\n", ingressObject.Name, NAMESPACE)
	return nil
}

func (d *Deployer) unDeployImage(appName string) error {
	err := d.kubernetesClient.
		AppsV1().
		Deployments(NAMESPACE).
		DeleteCollection(
			context.Background(),
			metav1.DeleteOptions{},
			metav1.ListOptions{LabelSelector: "app_name=" + ToK8sLabelValue(appName)},
		)
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}
	d.logger.LogInfoF("Deployment %q deleted in namespace %q\n", appName, NAMESPACE)
	return nil
}

func (d *Deployer) unExposeAppInternally(appName string) error {
	err := d.kubernetesClient.
		CoreV1().
		Services(NAMESPACE).
		Delete(
			context.Background(),
			ToK8sServiceName(appName),
			metav1.DeleteOptions{},
		)

	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}
	d.logger.LogInfoF("Service %q deleted in namespace %q\n", appName, NAMESPACE)
	return nil
}

func (d *Deployer) unExposeAppExternally(appName string) error {
	err := d.kubernetesClient.
		NetworkingV1().
		Ingresses(NAMESPACE).
		DeleteCollection(
			context.Background(),
			metav1.DeleteOptions{},
			metav1.ListOptions{LabelSelector: "app_name=" + ToK8sLabelValue(appName)},
		)
	if err != nil {
		return fmt.Errorf("failed to delete ingress: %w", err)
	}
	d.logger.LogInfoF("Ingress %q deleted in namespace %q\n", appName, NAMESPACE)
	return nil
}

func (d *Deployer) generateDeploymentObject(appName, imageURL string, labels map[string]string, envVars []v1Core.EnvVar) v1Apps.Deployment {
	return v1Apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   ToK8sDeploymentName(appName),
			Labels: labels,
		},
		Spec: v1Apps.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: v1Core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: v1Core.PodSpec{
					Containers: []v1Core.Container{
						{
							Name:  ToK8sContainerName(appName),
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

func (d *Deployer) generateServiceObject(namespace, appName string, labels map[string]string) v1Core.Service {
	return v1Core.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ToK8sServiceName(appName),
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: v1Core.ServiceSpec{
			Selector: labels,
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

func (d *Deployer) generateIngressObject(appName, host, serviceName string, labels map[string]string) networkingv1.Ingress {
	pathType := networkingv1.PathTypePrefix

	ingressClassName := "nginx"

	return networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ToK8sIngressName(appName),
			Labels:    labels,
			Namespace: NAMESPACE,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ingressClassName,
			TLS: []networkingv1.IngressTLS{
				{
					Hosts:      []string{host},
					SecretName: os.Getenv("TLS_WILDCARD_SECRET_NAME"),
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
