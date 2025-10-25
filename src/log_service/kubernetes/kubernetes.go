package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sort"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type KubernetesClient struct {
	Client *kubernetes.Clientset
}

const Namespace = "default"

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

func (kubernetesClient KubernetesClient) ReadPodLogs(appId string) (string, error) {
	pods, err := kubernetesClient.Client.
		CoreV1().
		Pods(Namespace).
		List(
			context.Background(),
			v1.ListOptions{
				LabelSelector: "app_id=" + appId,
			},
		)
	if err != nil {
		return "", fmt.Errorf("failed to list pods: %v", err)
	}

	// Sort Pods to get build logs before app logs
	sort.Slice(pods.Items, func(i, j int) bool {
		return pods.Items[i].CreationTimestamp.Unix() < pods.Items[j].CreationTimestamp.Unix()
	})

	logContent := ""

	for _, pod := range pods.Items {
		req := kubernetesClient.Client.
			CoreV1().
			Pods(Namespace).
			GetLogs(pod.Name, &corev1.PodLogOptions{})

		podLogs, err := req.Stream(context.Background())
		if err != nil {
			return "", fmt.Errorf("error opening log stream for pod %s: %v", pod.Name, err)
		}
		defer podLogs.Close()

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			return "", fmt.Errorf("error reading log stream for pod %s: %v", pod.Name, err)
		}

		logContent = logContent + buf.String()
		fmt.Printf("Logs for pod %s:\n%s\n", pod.Name, logContent)
	}

	return logContent, nil
}
