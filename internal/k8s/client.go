package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	clientset *kubernetes.Clientset
	namespace string
}

func NewClient(namespace string) (*Client, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("getting kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes client: %w", err)
	}

	return &Client{
		clientset: clientset,
		namespace: namespace,
	}, nil
}

func getKubeConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("getting home directory: %w", err)
		}
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("building config from kubeconfig: %w", err)
	}

	return config, nil
}

func (c *Client) CreatePod(ctx context.Context, pod *corev1.Pod) (*corev1.Pod, error) {
	return c.clientset.CoreV1().Pods(c.namespace).Create(ctx, pod, metav1.CreateOptions{})
}

func (c *Client) GetPod(ctx context.Context, name string) (*corev1.Pod, error) {
	return c.clientset.CoreV1().Pods(c.namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) DeletePod(ctx context.Context, name string) error {
	return c.clientset.CoreV1().Pods(c.namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (c *Client) GetPodLogs(ctx context.Context, name string) (string, error) {
	req := c.clientset.CoreV1().Pods(c.namespace).GetLogs(name, &corev1.PodLogOptions{})
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("opening log stream: %w", err)
	}
	defer stream.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, stream); err != nil {
		return "", fmt.Errorf("reading log stream: %w", err)
	}

	return buf.String(), nil
}

// StreamPodLogs opens a follow stream on the pod's logs and returns the reader.
// The caller is responsible for closing the returned ReadCloser.
func (c *Client) StreamPodLogs(ctx context.Context, name string) (io.ReadCloser, error) {
	req := c.clientset.CoreV1().Pods(c.namespace).GetLogs(name, &corev1.PodLogOptions{
		Follow: true,
	})

	stream, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("opening log stream: %w", err)
	}

	return stream, nil
}

// WaitForPodRunning blocks until the pod is in Running (or already completed) state.
func (c *Client) WaitForPodRunning(ctx context.Context, name string) error {
	watch, err := c.clientset.CoreV1().Pods(c.namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	})
	if err != nil {
		return fmt.Errorf("creating watch: %w", err)
	}
	defer watch.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-watch.ResultChan():
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}

			switch pod.Status.Phase {
			case corev1.PodRunning, corev1.PodSucceeded:
				return nil
			case corev1.PodFailed:
				return fmt.Errorf("pod %s failed before running", name)
			}
		}
	}
}

func (c *Client) CreateSecret(ctx context.Context, secret *corev1.Secret) (*corev1.Secret, error) {
	return c.clientset.CoreV1().Secrets(c.namespace).Create(ctx, secret, metav1.CreateOptions{})
}

func (c *Client) GetSecret(ctx context.Context, name string) (*corev1.Secret, error) {
	return c.clientset.CoreV1().Secrets(c.namespace).Get(ctx, name, metav1.GetOptions{})
}

func (c *Client) UpdateSecret(ctx context.Context, secret *corev1.Secret) (*corev1.Secret, error) {
	return c.clientset.CoreV1().Secrets(c.namespace).Update(ctx, secret, metav1.UpdateOptions{})
}

func (c *Client) EnsureNamespace(ctx context.Context, name string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				"name": name,
			},
		},
	}

	_, err := c.clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return fmt.Errorf("creating namespace %s: %w", name, err)
	}

	return nil
}

func (c *Client) WaitForPodCompletion(ctx context.Context, name string) error {
	watch, err := c.clientset.CoreV1().Pods(c.namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", name),
	})
	if err != nil {
		return fmt.Errorf("creating watch: %w", err)
	}
	defer watch.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-watch.ResultChan():
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				continue
			}

			if pod.Status.Phase == corev1.PodSucceeded {
				return nil
			}
			if pod.Status.Phase == corev1.PodFailed {
				return fmt.Errorf("pod %s failed", name)
			}
		}
	}
}
