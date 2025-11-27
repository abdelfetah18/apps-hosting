package buildexecutor

import (
	"context"
	"fmt"

	"apps-hosting.com/logging"
	"k8s.io/client-go/kubernetes"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KanikoExecutor struct {
	kubernetesClientset *kubernetes.Clientset
	logger              logging.ServiceLogger
}

func NewKanikoExecutor(kubernetesClientset *kubernetes.Clientset, logger logging.ServiceLogger) KanikoExecutor {
	return KanikoExecutor{
		kubernetesClientset: kubernetesClientset,
		logger:              logger,
	}
}

func (k *KanikoExecutor) Execute(srcContext, destination, appId, appName string) error {
	job := NewKanikoJob(srcContext, destination, appId, appName)

	_, err := k.kubernetesClientset.BatchV1().Jobs("default").Create(context.Background(), &job, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	watch, err := k.kubernetesClientset.BatchV1().Jobs("default").Watch(context.Background(), metav1.ListOptions{
		LabelSelector: "app_name=" + ToK8sLabelValue(appName),
	})

	if err != nil {
		return err
	}

	for event := range watch.ResultChan() {
		switch evt := event.Object.(type) {
		case *batchv1.Job:
			if evt.Status.Succeeded > 0 {
				k.logger.LogInfo("Job completed successfully")
				return nil
			}

			if evt.Status.Failed > 0 {
				k.logger.LogErrorF("job %s failed", job.Name)
				return fmt.Errorf("job %s failed", job.Name)
			}
		}
	}

	return err
}

func (k *KanikoExecutor) DeleteJobs(appName string) error {
	policy := metav1.DeletePropagationForeground

	return k.kubernetesClientset.
		BatchV1().
		Jobs("default").
		DeleteCollection(
			context.Background(),
			metav1.DeleteOptions{
				PropagationPolicy: &policy,
			},
			metav1.ListOptions{
				LabelSelector: "app_name=" + ToK8sLabelValue(appName),
			},
		)
}

func NewKanikoJob(srcContext, destination, appId, appName string) batchv1.Job {
	containerRestartPolicy := corev1.ContainerRestartPolicyNever

	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: ToK8sJobName(appName),
			Labels: map[string]string{
				"app_id":   ToK8sLabelValue(appId),
				"app_name": ToK8sLabelValue(appName),
			},
		},
		Spec: batchv1.JobSpec{
			// TTLSecondsAfterFinished: &ttlSeconds,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app_id":   ToK8sLabelValue(appId),
						"app_name": ToK8sLabelValue(appName),
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "kaniko",
							Image: "gcr.io/kaniko-project/executor:latest",
							Args: []string{
								fmt.Sprintf("--context=%s", srcContext),
								fmt.Sprintf("--destination=%s", destination),
								"--build-arg=START_CMD=start",
								"--build-arg=BUILD_CMD=build",
								"--insecure",
							},
							Env: []corev1.EnvVar{
								{Name: "AWS_ACCESS_KEY_ID", Value: "minioadmin"},
								{Name: "AWS_SECRET_ACCESS_KEY", Value: "minioadmin"},
								{Name: "S3_ENDPOINT", Value: "http://object-storage-minio:9000"},
								{Name: "S3_FORCE_PATH_STYLE", Value: "true"},
							},
							RestartPolicy: &containerRestartPolicy,
						},
					},
				},
			},
		},
	}
}
