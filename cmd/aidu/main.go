package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	namespace string
	image     string
	model     string
	logDir    string
)

var rootCmd = &cobra.Command{
	Use:   "aidu",
	Short: "AIDU - Artificial Intelligence Deployable Units",
	Long:  "Launch Claude Code sessions in Kubernetes pods to complete tasks autonomously.",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&namespace, "namespace", "aidu-sessions", "Kubernetes namespace")
	rootCmd.PersistentFlags().StringVar(&image, "image", "localhost:5050/aidu-claude:latest", "Worker container image")
	rootCmd.PersistentFlags().StringVar(&model, "model", "claude-opus-4-6", "Claude model to use")
	rootCmd.PersistentFlags().StringVar(&logDir, "log-dir", "./logs", "Directory to store pod logs")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
