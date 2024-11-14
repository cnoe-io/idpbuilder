package helpers

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	KubeConfigPath string
	scheme         *runtime.Scheme
)

func GetKubeConfigPath() string {
	if KubeConfigPath == "" {
		return filepath.Join(homedir.HomeDir(), ".kube", "config")
	} else {
		return KubeConfigPath
	}
}

func LoadKubeConfig() (*api.Config, error) {
	config, err := clientcmd.LoadFromFile(GetKubeConfigPath())
	if err != nil {
		return nil, fmt.Errorf("Failed to load kubeconfig file: %w", err)
	} else {
		return config, nil
	}
}

func GetKubeConfig() (*rest.Config, error) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", GetKubeConfigPath())
	if err != nil {
		return nil, fmt.Errorf("Error building kubeconfig: %w", err)
	}
	return kubeConfig, nil
}

func GetKubeClient(kubeConfig *rest.Config) (client.Client, error) {
	kubeClient, err := client.New(kubeConfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("Error creating kubernetes client: %w", err)
	}
	return kubeClient, nil
}
