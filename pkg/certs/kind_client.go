package certs

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// KindClient interface for interacting with Kind clusters
type KindClient interface {
	GetCurrentCluster() (string, error)
	ListClusters() ([]string, error)
	GetPods(ctx context.Context, namespace, labelSelector string) ([]string, error)
	CopyFromPod(ctx context.Context, podName, path string) ([]byte, error)
	ExecInPod(ctx context.Context, podName string, command []string) (string, error)
}

// KubectlKindClient implements KindClient using kubectl
type KubectlKindClient struct {
	kubectlPath string
	kubeconfig  string
}

// NewKubectlKindClient creates a new kubectl-based Kind client
func NewKubectlKindClient() (*KubectlKindClient, error) {
	// Find kubectl binary
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		return nil, fmt.Errorf("kubectl not found in PATH: %w", err)
	}

	// Verify kubectl is available
	if err := exec.Command(kubectlPath, "version", "--client", "--output=json").Run(); err != nil {
		return nil, fmt.Errorf("kubectl not working properly: %w", err)
	}

	// Setup kubeconfig path (use default if not set)
	kubeconfig := ""
	if kubeconfigFromEnv := getKubeconfigPath(); kubeconfigFromEnv != "" {
		kubeconfig = kubeconfigFromEnv
	}

	return &KubectlKindClient{
		kubectlPath: kubectlPath,
		kubeconfig:  kubeconfig,
	}, nil
}

// GetCurrentCluster returns the current Kind cluster name
func (c *KubectlKindClient) GetCurrentCluster() (string, error) {
	cmd := exec.Command("kind", "get", "clusters")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list Kind clusters: %w", err)
	}

	clusters := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(clusters) == 0 || (len(clusters) == 1 && clusters[0] == "") {
		return "", ErrNoKindCluster
	}

	// Return first cluster - the simplest approach for Kind clusters
	// Kind clusters are typically simple single-cluster environments
	return clusters[0], nil
}

// ListClusters returns all available Kind clusters
func (c *KubectlKindClient) ListClusters() ([]string, error) {
	cmd := exec.Command("kind", "get", "clusters")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}

	clusterList := strings.TrimSpace(string(output))
	if clusterList == "" {
		return []string{}, nil
	}

	return strings.Split(clusterList, "\n"), nil
}

// GetPods returns pods matching the label selector
func (c *KubectlKindClient) GetPods(ctx context.Context, namespace, labelSelector string) ([]string, error) {
	args := []string{"get", "pods", "-n", namespace}
	if labelSelector != "" {
		args = append(args, "-l", labelSelector)
	}
	args = append(args, "-o", "jsonpath={.items[*].metadata.name}")

	cmd := exec.CommandContext(ctx, c.kubectlPath, args...)
	if c.kubeconfig != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", c.kubeconfig))
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	podList := strings.TrimSpace(string(output))
	if podList == "" {
		return []string{}, nil
	}

	pods := strings.Fields(podList)
	return pods, nil
}

// CopyFromPod copies a file from a pod
func (c *KubectlKindClient) CopyFromPod(ctx context.Context, podName, path string) ([]byte, error) {
	// Use kubectl cp to extract file
	cmd := exec.CommandContext(ctx, c.kubectlPath, "cp",
		fmt.Sprintf("%s:%s", podName, path), "-")

	if c.kubeconfig != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", c.kubeconfig))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("kubectl cp failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// ExecInPod executes a command in a pod
func (c *KubectlKindClient) ExecInPod(ctx context.Context, podName string, command []string) (string, error) {
	args := append([]string{"exec", podName, "--"}, command...)
	cmd := exec.CommandContext(ctx, c.kubectlPath, args...)

	if c.kubeconfig != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("KUBECONFIG=%s", c.kubeconfig))
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("exec in pod failed: %w", err)
	}

	return string(output), nil
}

// getKubeconfigPath gets the kubeconfig path from environment or default location
func getKubeconfigPath() string {
	// Check KUBECONFIG environment variable first
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}

	// Default to ~/.kube/config
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".kube", "config")
	}

	return ""
}