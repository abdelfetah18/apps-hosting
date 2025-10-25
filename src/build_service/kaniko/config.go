package kaniko

import (
	"build/utils"
	"os"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewKanikoJob(appId, appName, imageName, repoFile string) batchv1.Job {
	registryURL := os.Getenv("REGISTRY_URL")

	return batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: utils.ToK8sJobName(appName),
			Labels: map[string]string{
				"app_id":   utils.ToK8sLabelValue(appId),
				"app_name": utils.ToK8sLabelValue(appName),
			},
		},
		Spec: batchv1.JobSpec{
			// TTLSecondsAfterFinished: &ttlSeconds,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app_id":   utils.ToK8sLabelValue(appId),
						"app_name": utils.ToK8sLabelValue(appName),
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  "kaniko",
							Image: registryURL + "executor:latest",
							Args: []string{
								"--context=http://build-service:8081/repos/" + repoFile,
								"--destination=" + imageName,
								"--insecure",
								"--skip-tls-verify",
								"--build-arg=START_CMD=start",
								"--build-arg=BUILD_CMD=build",
							},
						},
					},
				},
			},
		},
	}
}
