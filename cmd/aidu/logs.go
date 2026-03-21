package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/eduardvoiculescu/aidu/internal/k8s"
)

var logsCmd = &cobra.Command{
	Use:   "logs <pod-name>",
	Short: "Fetch and display logs from an AIDU pod",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		podName := args[0]

		client, err := k8s.NewClient(namespace)
		if err != nil {
			return fmt.Errorf("connecting to kubernetes: %w", err)
		}

		logs, err := client.GetPodLogs(cmd.Context(), podName)
		if err != nil {
			return fmt.Errorf("getting logs for pod %s: %w", podName, err)
		}

		fmt.Print(logs)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
}
