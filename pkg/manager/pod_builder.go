package manager

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/eduardvoiculescu/aidu/pkg/config"
)

type PodBuilder struct {
	namespace   string
	workerImage string
}

func NewPodBuilder(namespace, image string) *PodBuilder {
	if image == "" {
		image = "localhost:5050/aidu-claude:latest"
	}
	return &PodBuilder{
		namespace:   namespace,
		workerImage: image,
	}
}

func (pb *PodBuilder) BuildWorkerPod(taskID string, taskContext *config.TaskContext) *corev1.Pod {
	podName := fmt.Sprintf("worker-%s", taskID)
	configMapName := fmt.Sprintf("task-%s", taskID)
	pvcName := fmt.Sprintf("worker-%s-output", taskID)

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: pb.namespace,
			Labels: map[string]string{
				"app":     "aidu-worker",
				"manager": "aidu-orchestrator",
				"task-id": taskID,
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:    "claude-code",
					Image:   pb.workerImage,
					Command: []string{"/bin/bash", "-c"},
					Args: []string{
						`
set -e
echo "Starting AIDU worker for task: $(cat /task/task_id)"
echo "Complexity: $(cat /task/complexity)"
echo "---"

mkdir -p ~/.claude
cp /tmp/claude-creds/.credentials.json ~/.claude/.credentials.json
chmod 600 ~/.claude/.credentials.json

cd /workspace

cat /task/task.md | claude --print > /output/result.md 2> /output/logs.txt
EXIT_CODE=$?

echo $EXIT_CODE > /output/exit_code
echo "---"
echo "Worker completed with exit code: $EXIT_CODE"

exit $EXIT_CODE
`,
					},
					Env: []corev1.EnvVar{
						{
							Name:  "HOME",
							Value: "/home/claude",
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "task-context",
							MountPath: "/task",
							ReadOnly:  true,
						},
						{
							Name:      "output",
							MountPath: "/output",
						},
						{
							Name:      "workspace",
							MountPath: "/workspace",
						},
						{
							Name:      "claude-credentials",
							MountPath: "/tmp/claude-creds",
							ReadOnly:  true,
						},
					},
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("4Gi"),
							corev1.ResourceCPU:    resource.MustParse("2"),
						},
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("2Gi"),
							corev1.ResourceCPU:    resource.MustParse("1"),
						},
					},
					ImagePullPolicy: corev1.PullNever,
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "task-context",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: configMapName,
							},
						},
					},
				},
				{
					Name: "output",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvcName,
						},
					},
				},
				{
					Name: "workspace",
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				{
					Name: "claude-credentials",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "claude-credentials",
						},
					},
				},
			},
		},
	}
}

func (pb *PodBuilder) BuildConfigMap(taskID string, data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("task-%s", taskID),
			Namespace: pb.namespace,
			Labels: map[string]string{
				"app":     "aidu-worker",
				"task-id": taskID,
			},
		},
		Data: data,
	}
}

func (pb *PodBuilder) BuildPVC(taskID string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("worker-%s-output", taskID),
			Namespace: pb.namespace,
			Labels: map[string]string{
				"app":     "aidu-worker",
				"task-id": taskID,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		},
	}
}
