package runner

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const credentialsSecretName = "claude-credentials"
const credentialsSecretKey = ".credentials.json"

func buildPod(name, namespace, image, task, model string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app":      "aidu-worker",
				"aidu-pod": name,
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:    "claude-code",
					Image:   image,
					Command: []string{"/bin/bash", "-c"},
					Args: []string{
						`mkdir -p ~/.claude && ` +
							`cp /tmp/claude-creds/.credentials.json ~/.claude/.credentials.json && ` +
							`chmod 600 ~/.claude/.credentials.json && ` +
							`exec claude -p "$AIDU_TASK" --model "$AIDU_MODEL" --output-format stream-json --verbose`,
					},
					Env: []corev1.EnvVar{
						{
							Name:  "AIDU_TASK",
							Value: task,
						},
						{
							Name:  "AIDU_MODEL",
							Value: model,
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "claude-credentials",
							MountPath: "/tmp/claude-creds",
							ReadOnly:  true,
						},
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("512Mi"),
							corev1.ResourceCPU:    resource.MustParse("500m"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("4Gi"),
							corev1.ResourceCPU:    resource.MustParse("2"),
						},
					},
					ImagePullPolicy: corev1.PullNever,
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "claude-credentials",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: credentialsSecretName,
						},
					},
				},
			},
		},
	}
}
