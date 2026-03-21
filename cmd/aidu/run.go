package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/eduardvoiculescu/aidu/internal/k8s"
	"github.com/eduardvoiculescu/aidu/internal/runner"
)

var runCmd = &cobra.Command{
	Use:   "run <task>",
	Short: "Launch a Claude Code pod to execute a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		task := args[0]

		client, err := k8s.NewClient(namespace)
		if err != nil {
			return fmt.Errorf("connecting to kubernetes: %w", err)
		}

		if err := client.EnsureNamespace(cmd.Context(), namespace); err != nil {
			return fmt.Errorf("ensuring namespace: %w", err)
		}

		r := &runner.Runner{
			K8sClient: client,
			Namespace: namespace,
			Image:     image,
			Model:     model,
			LogDir:    logDir,
		}

		podName, err := r.Run(cmd.Context(), task)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Task failed: %v\n", err)
			if podName != "" {
				fmt.Fprintf(os.Stderr, "Check logs: kubectl logs %s -n %s\n", podName, namespace)
			}
			os.Exit(1)
		}

		fmt.Println("Task completed successfully!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
