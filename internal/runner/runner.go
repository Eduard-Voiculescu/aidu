package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/eduardvoiculescu/aidu/internal/k8s"
)

type Runner struct {
	K8sClient *k8s.Client
	Namespace string
	Image     string
	Model     string
	LogDir    string
}

func (r *Runner) Run(ctx context.Context, task string) (string, error) {
	podName := fmt.Sprintf("aidu-%d", time.Now().Unix())

	pod := buildPod(podName, r.Namespace, r.Image, task, r.Model)

	if _, err := r.K8sClient.CreatePod(ctx, pod); err != nil {
		return "", fmt.Errorf("creating pod: %w", err)
	}

	fmt.Printf("Pod created: %s\n", podName)
	fmt.Println("Waiting for pod to start...")

	if err := r.K8sClient.WaitForPodRunning(ctx, podName); err != nil {
		return podName, fmt.Errorf("waiting for pod to start: %w", err)
	}

	logPath, err := r.prepareLogFile(podName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: cannot create log file: %v\n", err)
	}

	fmt.Println("--- streaming output ---")

	streamErr := r.streamAndParseLogs(ctx, podName, logPath)

	fmt.Println()

	completionErr := r.K8sClient.WaitForPodCompletion(ctx, podName)

	if logPath != "" {
		fmt.Printf("--- logs saved: %s ---\n", logPath)
	}

	if err := r.K8sClient.DeletePod(ctx, podName); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to clean up pod: %v\n", err)
	}

	if streamErr != nil {
		return podName, fmt.Errorf("streaming logs: %w", streamErr)
	}
	if completionErr != nil {
		return podName, fmt.Errorf("pod execution failed: %w", completionErr)
	}

	return podName, nil
}

func (r *Runner) streamAndParseLogs(ctx context.Context, podName, logPath string) error {
	stream, err := r.K8sClient.StreamPodLogs(ctx, podName)
	if err != nil {
		return fmt.Errorf("opening log stream: %w", err)
	}
	defer stream.Close()

	parsedText, parseErr := parseStreamJSON(stream, os.Stdout)

	if logPath != "" && parsedText != "" {
		if writeErr := os.WriteFile(logPath, []byte(parsedText), 0644); writeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: cannot write log file: %v\n", writeErr)
		}
	}

	return parseErr
}

func (r *Runner) prepareLogFile(podName string) (string, error) {
	if err := os.MkdirAll(r.LogDir, 0755); err != nil {
		return "", fmt.Errorf("creating log directory: %w", err)
	}
	return filepath.Join(r.LogDir, podName+".log"), nil
}
