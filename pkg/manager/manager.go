package manager

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/eduardvoiculescu/aidu/internal/k8s"
	"github.com/eduardvoiculescu/aidu/pkg/analyzer"
	"github.com/eduardvoiculescu/aidu/pkg/config"
	"github.com/eduardvoiculescu/aidu/pkg/types"
	"github.com/eduardvoiculescu/aidu/pkg/validator"
)

type ManagerConfig struct {
	Namespace          string
	MaxRetries         int
	Timeout            time.Duration
	ValidationScore    float64
	WorkerImage        string
	AgentsMDPath       string
	OutputDir          string
}

type Manager struct {
	config     *ManagerConfig
	k8sClient  *k8s.Client
	analyzer   *analyzer.Analyzer
	validator  *validator.Validator
	podBuilder *PodBuilder
}

func NewManager(cfg *ManagerConfig) (*Manager, error) {
	if cfg.Namespace == "" {
		cfg.Namespace = "aidu-sessions"
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Minute
	}
	if cfg.ValidationScore == 0 {
		cfg.ValidationScore = 0.75
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = "./output"
	}

	k8sClient, err := k8s.NewClient(cfg.Namespace)
	if err != nil {
		return nil, fmt.Errorf("creating k8s client: %w", err)
	}

	return &Manager{
		config:     cfg,
		k8sClient:  k8sClient,
		analyzer:   analyzer.New(),
		validator:  validator.New(cfg.ValidationScore),
		podBuilder: NewPodBuilder(cfg.Namespace, cfg.WorkerImage),
	}, nil
}

func (m *Manager) ExecuteTask(ctx context.Context, description string) (*types.TaskResult, error) {
	log.Printf("Received task: %s", description)

	complexity := m.analyzer.AnalyzeTask(description)
	log.Printf("Task complexity: %s", complexity)

	agentsCfg, err := config.LoadAgentsMD(m.config.AgentsMDPath)
	if err != nil {
		return nil, fmt.Errorf("loading AGENTS.md: %w", err)
	}
	log.Printf("Loaded agent context from: %s", agentsCfg.Path)

	task := &types.Task{
		ID:          generateTaskID(),
		Description: description,
		Complexity:  complexity,
		AgentsMD:    agentsCfg.Content,
		CreatedAt:   time.Now(),
		RetryCount:  0,
	}

	for attempt := 0; attempt <= m.config.MaxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retry attempt %d/%d", attempt, m.config.MaxRetries)
			task.RetryCount = attempt
			task.ID = generateTaskID()
		}

		result, err := m.executeAttempt(ctx, task)
		if err != nil {
			log.Printf("Attempt %d failed: %v", attempt, err)
			if err := m.cleanup(ctx, task.ID); err != nil {
				log.Printf("Warning: cleanup after error failed: %v", err)
			}
			if attempt == m.config.MaxRetries {
				return nil, fmt.Errorf("max retries exceeded: %w", err)
			}
			continue
		}

		if result.Validation.Valid {
			log.Printf("Task completed successfully! Score: %.2f", result.Validation.Score)
			if err := m.cleanup(ctx, task.ID); err != nil {
				log.Printf("Warning: cleanup after success failed: %v", err)
			}
			return result, nil
		}

		if !result.Validation.ShouldRetry {
			return result, nil
		}

		task.Feedback = result.Validation.Feedback
		log.Printf("Validation failed (score: %.2f). Retrying with feedback...", result.Validation.Score)

		if err := m.cleanup(ctx, task.ID); err != nil {
			log.Printf("Warning: cleanup failed: %v", err)
		}
	}

	return nil, fmt.Errorf("task failed after %d attempts", m.config.MaxRetries)
}

func (m *Manager) executeAttempt(ctx context.Context, task *types.Task) (*types.TaskResult, error) {
	ctx, cancel := context.WithTimeout(ctx, m.config.Timeout)
	defer cancel()

	log.Printf("Creating worker for task %s", task.ID)

	taskContext := config.BuildTaskContext(task, task.AgentsMD)

	configMap := m.podBuilder.BuildConfigMap(task.ID, taskContext.ToConfigMapData())
	if _, err := m.k8sClient.CreateConfigMap(ctx, configMap); err != nil {
		return nil, fmt.Errorf("creating configmap: %w", err)
	}

	pvc := m.podBuilder.BuildPVC(task.ID)
	if _, err := m.k8sClient.CreatePVC(ctx, pvc); err != nil {
		return nil, fmt.Errorf("creating pvc: %w", err)
	}

	pod := m.podBuilder.BuildWorkerPod(task.ID, taskContext)
	if _, err := m.k8sClient.CreatePod(ctx, pod); err != nil {
		return nil, fmt.Errorf("creating pod: %w", err)
	}

	podName := fmt.Sprintf("worker-%s", task.ID)
	log.Printf("Waiting for worker pod %s to start...", podName)

	if err := m.k8sClient.WaitForPodReady(ctx, podName); err != nil {
		return nil, fmt.Errorf("waiting for pod ready: %w", err)
	}

	log.Printf("Worker pod running, waiting for completion...")
	startTime := time.Now()

	if err := m.k8sClient.WaitForPodCompletion(ctx, podName); err != nil {
		return nil, fmt.Errorf("waiting for pod completion: %w", err)
	}

	executionTime := time.Since(startTime).Seconds()
	log.Printf("Worker completed in %.2f seconds", executionTime)

	outputPath := fmt.Sprintf("/tmp/aidu-output/%s", task.ID)
	if err := m.retrieveOutput(ctx, task.ID, outputPath); err != nil {
		return nil, fmt.Errorf("retrieving output: %w", err)
	}

	validation := m.validator.ValidateOutput(outputPath, task)

	result := &types.TaskResult{
		TaskID:        task.ID,
		WorkerID:      podName,
		Success:       validation.Valid,
		OutputPath:    outputPath,
		Validation:    validation,
		ExecutionTime: executionTime,
	}

	return result, nil
}

func (m *Manager) retrieveOutput(ctx context.Context, taskID, destPath string) error {
	log.Printf("Retrieving output to %s", destPath)

	pvcName := fmt.Sprintf("worker-%s-output", taskID)

	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	copyPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("copy-%s", taskID),
			Namespace: m.config.Namespace,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:    "copy",
					Image:   "busybox",
					Command: []string{"sh", "-c", "sleep 300"},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "output",
							MountPath: "/output",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "output",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvcName,
						},
					},
				},
			},
		},
	}

	if _, err := m.k8sClient.CreatePod(ctx, copyPod); err != nil {
		return fmt.Errorf("creating copy pod: %w", err)
	}
	defer m.k8sClient.DeletePod(ctx, copyPod.Name)

	if err := m.k8sClient.WaitForPodReady(ctx, copyPod.Name); err != nil {
		return fmt.Errorf("waiting for copy pod: %w", err)
	}

	files := []string{"result.md", "logs.txt", "exit_code"}
	for _, file := range files {
		srcPath := fmt.Sprintf("%s:/output/%s", copyPod.Name, file)
		dstPath := fmt.Sprintf("%s/%s", destPath, file)

		cmd := exec.CommandContext(ctx, "kubectl", "cp",
			"-n", m.config.Namespace,
			srcPath, dstPath)
		if err := cmd.Run(); err != nil {
			log.Printf("Warning: failed to copy %s: %v", file, err)
		}
	}

	return nil
}

func (m *Manager) cleanup(ctx context.Context, taskID string) error {
	podName := fmt.Sprintf("worker-%s", taskID)
	configMapName := fmt.Sprintf("task-%s", taskID)
	pvcName := fmt.Sprintf("worker-%s-output", taskID)

	if err := m.k8sClient.DeletePod(ctx, podName); err != nil {
		log.Printf("Warning: failed to delete pod: %v", err)
	}

	if err := m.k8sClient.DeleteConfigMap(ctx, configMapName); err != nil {
		log.Printf("Warning: failed to delete configmap: %v", err)
	}

	if err := m.k8sClient.DeletePVC(ctx, pvcName); err != nil {
		log.Printf("Warning: failed to delete pvc: %v", err)
	}

	return nil
}

func generateTaskID() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}
