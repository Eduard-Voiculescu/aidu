package runner

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestBuildPod(t *testing.T) {
	tests := []struct {
		name      string
		podName   string
		namespace string
		image     string
		task      string
		model     string
	}{
		{
			name:      "basic task",
			podName:   "aidu-123",
			namespace: "aidu-sessions",
			image:     "localhost:5050/aidu-claude:latest",
			task:      "What is 2+2?",
			model:     "claude-opus-4-6",
		},
		{
			name:      "complex task with special chars",
			podName:   "aidu-456",
			namespace: "custom-ns",
			image:     "my-registry/claude:v1",
			task:      `Build a REST API with "JWT auth" & rate limiting`,
			model:     "claude-sonnet-4-6",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pod := buildPod(tt.podName, tt.namespace, tt.image, tt.task, tt.model)

			if pod.Name != tt.podName {
				t.Errorf("pod name = %q, want %q", pod.Name, tt.podName)
			}

			if pod.Namespace != tt.namespace {
				t.Errorf("namespace = %q, want %q", pod.Namespace, tt.namespace)
			}

			if pod.Labels["app"] != "aidu-worker" {
				t.Errorf("label app = %q, want %q", pod.Labels["app"], "aidu-worker")
			}

			if pod.Spec.RestartPolicy != corev1.RestartPolicyNever {
				t.Errorf("restart policy = %q, want Never", pod.Spec.RestartPolicy)
			}

			container := pod.Spec.Containers[0]

			if container.Image != tt.image {
				t.Errorf("image = %q, want %q", container.Image, tt.image)
			}

			if container.Command[0] != "/bin/bash" {
				t.Errorf("command = %q, want /bin/bash", container.Command[0])
			}

			shellScript := container.Args[0]
			if !strings.Contains(shellScript, "claude -p") {
				t.Error("shell script does not contain 'claude -p'")
			}
			if !strings.Contains(shellScript, ".credentials.json") {
				t.Error("shell script does not copy credentials")
			}

			envMap := map[string]string{}
			for _, env := range container.Env {
				envMap[env.Name] = env.Value
			}

			if envMap["AIDU_TASK"] != tt.task {
				t.Errorf("AIDU_TASK = %q, want %q", envMap["AIDU_TASK"], tt.task)
			}
			if envMap["AIDU_MODEL"] != tt.model {
				t.Errorf("AIDU_MODEL = %q, want %q", envMap["AIDU_MODEL"], tt.model)
			}

			if len(container.VolumeMounts) != 1 {
				t.Fatalf("volume mounts count = %d, want 1", len(container.VolumeMounts))
			}
			if container.VolumeMounts[0].MountPath != "/tmp/claude-creds" {
				t.Errorf("volume mount path = %q, want /tmp/claude-creds", container.VolumeMounts[0].MountPath)
			}

			if len(pod.Spec.Volumes) != 1 {
				t.Fatalf("volumes count = %d, want 1", len(pod.Spec.Volumes))
			}
			vol := pod.Spec.Volumes[0]
			if vol.Secret == nil || vol.Secret.SecretName != credentialsSecretName {
				t.Errorf("volume secret name = %v, want %q", vol.Secret, credentialsSecretName)
			}

			if container.ImagePullPolicy != corev1.PullNever {
				t.Errorf("image pull policy = %q, want Never", container.ImagePullPolicy)
			}
		})
	}
}
