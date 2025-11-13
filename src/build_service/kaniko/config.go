package kaniko

import (
	"build/utils"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewKanikoJob(appId, appName, imageName, repoFile string) batchv1.Job {
	containerRestartPolicy := corev1.ContainerRestartPolicyNever

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
							Image: "gcr.io/kaniko-project/executor:latest",
							Args: []string{
								"--context=s3://apps-source/" + repoFile,
								"--destination=" + imageName,
								"--build-arg=START_CMD=start",
								"--build-arg=BUILD_CMD=build",
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
