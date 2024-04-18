package k8s

import (
	"embed"
	"os"

	"github.com/cnoe-io/idpbuilder/pkg/util"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
