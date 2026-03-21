package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/eduardvoiculescu/aidu/internal/k8s"
)

const (
	credentialsSecretName = "claude-credentials"
	credentialsSecretKey  = ".credentials.json"
)

var credentialsPath string

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Store Claude credentials as a Kubernetes secret for AIDU workers",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := credentialsPath
		if path == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("getting home directory: %w", err)
			}
			path = filepath.Join(home, ".claude", ".credentials.json")
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading credentials from %s: %w", path, err)
		}

		if len(data) == 0 {
			return fmt.Errorf("credentials file is empty")
		}

		fmt.Printf("Read credentials from: %s\n", path)

		client, err := k8s.NewClient(namespace)
		if err != nil {
			return fmt.Errorf("connecting to kubernetes: %w", err)
		}

		ctx := cmd.Context()

		if err := client.EnsureNamespace(ctx, namespace); err != nil {
			return fmt.Errorf("ensuring namespace: %w", err)
		}

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      credentialsSecretName,
				Namespace: namespace,
			},
			Data: map[string][]byte{
				credentialsSecretKey: data,
			},
		}

		existing, err := client.GetSecret(ctx, credentialsSecretName)
		if err != nil {
			if _, err := client.CreateSecret(ctx, secret); err != nil {
				return fmt.Errorf("creating secret: %w", err)
			}
			fmt.Println("Credentials stored successfully.")
			return nil
		}

		existing.Data = secret.Data
		if _, err := client.UpdateSecret(ctx, existing); err != nil {
			return fmt.Errorf("updating secret: %w", err)
		}

		fmt.Println("Credentials updated successfully.")
		return nil
	},
}

func init() {
	setupCmd.Flags().StringVar(&credentialsPath, "credentials", "", "Path to .credentials.json (default: ~/.claude/.credentials.json)")
	rootCmd.AddCommand(setupCmd)
}
