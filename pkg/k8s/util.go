package k8s

import (
	"embed"
	"github.com/cnoe-io/idpbuilder/pkg/util/files"
	"github.com/cnoe-io/idpbuilder/pkg/util/fs"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func BuildCustomizedManifests(filePath, fsPath string, resourceFS embed.FS, scheme *runtime.Scheme, templateData any) ([][]byte, error) {
	rawResources, err := fs.ConvertFSToBytes(resourceFS, fsPath, templateData)
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
	rawResources, err := fs.ConvertFSToBytes(resourceFS, fsPath, templateData)
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

	rendered, err := files.ApplyTemplate(customBS, templateData)
	if err != nil {
		return nil, nil, err
	}

	return ConvertYamlToObjectsWithOverride(scheme, originalFiles, rendered)
}

func DeploymentImages(deployment appsv1.Deployment) []string {
	images := []string{}
	for _, c := range deployment.Spec.Template.Spec.Containers {
		images = append(images, c.Image)
	}
	for _, c := range deployment.Spec.Template.Spec.InitContainers {
		images = append(images, c.Image)
	}

	return images
}

func StatefulSetImages(statefulset appsv1.StatefulSet) []string {
	images := []string{}
	for _, c := range statefulset.Spec.Template.Spec.Containers {
		images = append(images, c.Image)
	}
	for _, c := range statefulset.Spec.Template.Spec.InitContainers {
		images = append(images, c.Image)
	}

	return images
}

func JobImages(job batchv1.Job) []string {
	images := []string{}
	for _, c := range job.Spec.Template.Spec.Containers {
		images = append(images, c.Image)
	}
	for _, c := range job.Spec.Template.Spec.InitContainers {
		images = append(images, c.Image)
	}

	return images
}

func DaemonSetImages(daemonset appsv1.DaemonSet) []string {
	images := []string{}
	for _, c := range daemonset.Spec.Template.Spec.Containers {
		images = append(images, c.Image)
	}
	for _, c := range daemonset.Spec.Template.Spec.InitContainers {
		images = append(images, c.Image)
	}

	return images
}

func ObjectsImages(objects []client.Object) []string {
	images := []string{}
	for _, o := range objects {
		switch v := o.(type) {
		case *appsv1.DaemonSet:
			images = append(images, DaemonSetImages(*v)...)
		case *appsv1.StatefulSet:
			images = append(images, StatefulSetImages(*v)...)
		case *appsv1.Deployment:
			images = append(images, DeploymentImages(*v)...)
		case *batchv1.Job:
			images = append(images, JobImages(*v)...)
		default:
		}
	}
	return images
}
