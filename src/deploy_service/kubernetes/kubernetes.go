package kubernetes

import (
	"context"
	"deploy/utils"
	"fmt"
	"path/filepath"

	"apps-hosting.com/logging"

	v1Apps "k8s.io/api/apps/v1"
	v1Core "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type KubernetesClient struct {
	Client *kubernetes.Clientset
	Logger logging.ServiceLogger
}

func NewKubernetesClient(config *rest.Config) KubernetesClient {
	clientset, _ := kubernetes.NewForConfig(config)
	return KubernetesClient{
		Client: clientset,
	}
}

func GetKubernetesConfigFromEnv() (*rest.Config, error) {
	kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kube config: %v", err)
	}

	return config, nil
}

func (kubernetesClient KubernetesClient) DeployImage(namespace string, imageURL string, deployment v1Apps.Deployment) error {
	deploymentsClient := kubernetesClient.Client.AppsV1().Deployments(namespace)
	_, err := deploymentsClient.Create(context.Background(), &deployment, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to deploy the image: %v", err)
	}

	kubernetesClient.Logger.LogInfoF("Deployment created successfully in namespace %s with image: %s", namespace, imageURL)
	return nil
}

func (kubernetesClient KubernetesClient) CreateNamespace(namespace string) (*v1Core.Namespace, error) {
	return kubernetesClient.Client.CoreV1().Namespaces().Create(context.Background(), &v1Core.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}, metav1.CreateOptions{})
}

func (kubernetesClient KubernetesClient) CreateServiceForDeployment(namespace string, service v1Core.Service) (*v1Core.Service, error) {
	return kubernetesClient.
		Client.CoreV1().
		Services(namespace).
		Create(context.TODO(), &service, metav1.CreateOptions{})
}

func (kubernetesClient KubernetesClient) CreateIngress(namespace, ingressName string, ingress networkingv1.Ingress) error {
	_, err := kubernetesClient.Client.NetworkingV1().Ingresses(namespace).Create(context.TODO(), &ingress, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create ingress: %w", err)
	}

	kubernetesClient.Logger.LogInfoF("Ingress %q created in namespace %q\n", ingressName, namespace)
	return nil
}

func (kubernetesClient KubernetesClient) DeleteDeployment(namespace, appName string) error {
	err := kubernetesClient.Client.
		AppsV1().
		Deployments(namespace).
		DeleteCollection(
			context.Background(),
			metav1.DeleteOptions{},
			metav1.ListOptions{LabelSelector: "app_name=" + utils.ToK8sLabelValue(appName)},
		)

	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}

	kubernetesClient.Logger.LogInfoF("Deployment %q deleted in namespace %q\n", appName, namespace)
	return nil
}

func (kubernetesClient KubernetesClient) DeleteServiceForDeployment(namespace, appName string) error {
	err := kubernetesClient.Client.
		CoreV1().
		Services(namespace).
		Delete(
			context.Background(),
			utils.ToK8sServiceName(appName),
			metav1.DeleteOptions{},
		)

	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	kubernetesClient.Logger.LogInfoF("Service %q deleted in namespace %q\n", appName, namespace)
	return nil
}

func (kubernetesClient KubernetesClient) DeleteIngress(namespace, appName string) error {
	err := kubernetesClient.Client.
		NetworkingV1().
		Ingresses(namespace).
		DeleteCollection(
			context.Background(),
			metav1.DeleteOptions{},
			metav1.ListOptions{LabelSelector: "app_name=" + utils.ToK8sLabelValue(appName)},
		)
	if err != nil {
		return fmt.Errorf("failed to delete ingress: %w", err)
	}

	kubernetesClient.Logger.LogInfoF("Ingress %q deleted in namespace %q\n", appName, namespace)
	return nil
}
