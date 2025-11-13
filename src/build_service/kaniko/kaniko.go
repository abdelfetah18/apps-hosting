package kaniko

import (
	"build/utils"
	"context"
	"fmt"

	"apps-hosting.com/logging"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type KanikoBuilder struct {
	KubernetesClientset *kubernetes.Clientset
	Logger              logging.ServiceLogger
}

func NewKanikoBuilder(kubernetesClientset *kubernetes.Clientset, logger logging.ServiceLogger) KanikoBuilder {
	return KanikoBuilder{
		KubernetesClientset: kubernetesClientset,
		Logger:              logger,
	}
}

func (kanikoBuilder *KanikoBuilder) RunKanikoBuild(repoFile, appId, appName, imageName string) error {
	job := NewKanikoJob(appId, appName, imageName, repoFile)

	_, err := kanikoBuilder.KubernetesClientset.BatchV1().Jobs("default").Create(context.TODO(), &job, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	watch, err := kanikoBuilder.KubernetesClientset.BatchV1().Jobs("default").Watch(context.Background(), metav1.ListOptions{
		LabelSelector: "app_name=" + utils.ToK8sLabelValue(appName),
	})

	if err != nil {
		return err
	}

	for event := range watch.ResultChan() {
		switch evt := event.Object.(type) {
		case *batchv1.Job:
			if evt.Status.Succeeded > 0 {
				kanikoBuilder.Logger.LogInfo("Job completed successfully")
				return nil
			}

			if evt.Status.Failed > 0 {
				kanikoBuilder.Logger.LogErrorF("job %s failed", job.Name)
				return fmt.Errorf("job %s failed", job.Name)
			}
		}
	}

	return err
}

func (kanikoBuilder *KanikoBuilder) DeleteJobs(appName string) error {
	return kanikoBuilder.KubernetesClientset.
		BatchV1().
		Jobs("default").
		DeleteCollection(
			context.Background(),
			metav1.DeleteOptions{},
			metav1.ListOptions{
				LabelSelector: "app_name=" + utils.ToK8sLabelValue(appName),
			},
		)
}
