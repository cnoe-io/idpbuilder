package k8s

import (
	"embed"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/cnoe-io/idpbuilder/pkg/util"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	setupLog = ctrl.Log.WithName("k8s")
)

func BuildCustomizedManifests(filePath, fsPath string, resourceFS embed.FS, scheme *runtime.Scheme, templateData any) ([][]byte, error) {
	rawResources, err := util.ConvertFSToBytes(resourceFS, fsPath, templateData)
	if err != nil {
		return nil, err
	}

	if filePath == "" {
		return rawResources, nil
	}

	bs, _, err := applyOverrides(filePath, rawResources, scheme, templateData)
	if err != nil {
		return nil, err
	}

	return bs, nil
}

func BuildCustomizedObjects(filePath, fsPath string, resourceFS embed.FS, scheme *runtime.Scheme, templateData any) ([]client.Object, error) {
	rawResources, err := util.ConvertFSToBytes(resourceFS, fsPath, templateData)
	if err != nil {
		return nil, err
	}

	if filePath == "" {
		return ConvertRawResourcesToObjects(scheme, rawResources)
	}

	_, objs, err := applyOverrides(filePath, rawResources, scheme, templateData)
	if err != nil {
		return nil, err
	}

	return objs, nil
}

func applyOverrides(filePath string, originalFiles [][]byte, scheme *runtime.Scheme, templateData any) ([][]byte, []client.Object, error) {
	customBS, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}

	rendered, err := util.ApplyTemplate(customBS, templateData)
	if err != nil {
		return nil, nil, err
	}

	return ConvertYamlToObjectsWithOverride(scheme, originalFiles, rendered)
}

func GetKubeConfig(kubeConfigPath ...string) (*rest.Config, error) {
	// Set default path if no path is provided
	path := filepath.Join(homedir.HomeDir(), ".kube", "config")

	if len(kubeConfigPath) > 0 {
		path = kubeConfigPath[0]
	}

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		setupLog.Error(err, "Error building kubeconfig from kind cluster")
		return nil, err
	}
	return kubeConfig, nil
}

func GetKubeClient(kubeConfigPath ...string) (client.Client, error) {
	kubeCfg, err := GetKubeConfig(kubeConfigPath...)
	if err != nil {
		setupLog.Error(err, "Error getting kubeconfig")
		return nil, err
	}
	kubeClient, err := client.New(kubeCfg, client.Options{Scheme: GetScheme()})
	if err != nil {
		setupLog.Error(err, "Error creating kubernetes client")
		return nil, err
	}
	return kubeClient, nil
}
