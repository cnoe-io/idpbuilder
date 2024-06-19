package k8s

import (
	"bytes"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kyaml/kio"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

type ConversionError struct {
	rtObject runtime.Object
}

func (e *ConversionError) Error() string {
	return fmt.Sprintf("Failed to convert object %q", e.rtObject.GetObjectKind().GroupVersionKind().String())
}

func ConvertYamlToObjects(scheme *runtime.Scheme, objYamls []byte) ([]client.Object, error) {
	decode := serializer.NewCodecFactory(scheme).UniversalDeserializer().Decode

	var k8sObjects []client.Object

	for _, objYaml := range bytes.Split(objYamls, []byte{'\n', '-', '-', '-', '\n'}) {
		if len(objYaml) == 0 {
			continue
		}

		rtObject, _, err := decode(objYaml, nil, nil)
		if err != nil {
			return nil, err
		}
		object, ok := rtObject.(client.Object)
		if !ok {
			return nil, &ConversionError{rtObject: rtObject}
		}
		k8sObjects = append(k8sObjects, object)
	}
	return k8sObjects, nil
}

func ConvertRawResourcesToObjects(scheme *runtime.Scheme, rawResources [][]byte) ([]client.Object, error) {
	var ret []client.Object
	for _, resources := range rawResources {
		objs, err := ConvertYamlToObjects(scheme, resources)
		if err != nil {
			return nil, err
		}
		ret = append(ret, objs...)
	}
	return ret, nil
}

// replace k8s objects in given YAML doc with override objects. returns built yaml file and objects
func ConvertYamlToObjectsWithOverride(scheme *runtime.Scheme, originalFiles [][]byte, overrideYamls []byte) ([][]byte, []client.Object, error) {

	overrides, err := kio.FromBytes(overrideYamls)
	if err != nil {
		return nil, nil, err
	}

	overrideMap := make(map[string]*kyaml.RNode)
	order := make([]string, 0, len(overrides))
	for i := range overrides {
		o := overrides[i]
		id := GetObjectIdentifier(o)
		overrideMap[id] = o
		order = append(order, id)
	}

	outYaml := make([][]byte, len(originalFiles))
	outObjs := make([]client.Object, 0, 10)

	for i := range originalFiles {
		originalFile := originalFiles[i]
		originals, oErr := kio.FromBytes(originalFile)
		if oErr != nil {
			return nil, nil, oErr
		}

		for j := range originals {
			id := GetObjectIdentifier(originals[j])

			o, ok := overrideMap[id]
			if ok {
				// found an object that needs to be overridden. update manifest and remove from our map.
				originals[j].SetYNode(o.YNode())
				delete(overrideMap, id)
			}
		}

		manifest, oErr := kio.StringAll(originals)
		if oErr != nil {
			return nil, nil, fmt.Errorf("converting overridden manifest to string: %w", oErr)
		}

		objs, oErr := ConvertYamlToObjects(scheme, []byte(manifest))
		if oErr != nil {
			return nil, nil, fmt.Errorf("converting overridden manifest to k8s objects: %w", oErr)
		}
		outObjs = append(outObjs, objs...)
		outYaml[i] = []byte(manifest)
	}

	// if there are objects that are not overriding any original object, create a new file and add them to it.
	if len(overrideMap) != 0 {
		// must preserve original order
		n := make([]*kyaml.RNode, 0, len(overrideYamls))
		for i := range order {
			o, ok := overrideMap[order[i]]
			if ok {
				n = append(n, o)
			}
		}

		manifest, err := kio.StringAll(n)
		if err != nil {
			return nil, nil, fmt.Errorf("converting overridden manifest to string: %w", err)
		}

		objs, oErr := ConvertYamlToObjects(scheme, []byte(manifest))
		if oErr != nil {
			return nil, nil, fmt.Errorf("converting overridden manifest to k8s objects: %w", oErr)
		}

		outObjs = append(outObjs, objs...)
		outYaml = append(outYaml, []byte(manifest))
	}

	return outYaml, outObjs, nil
}

func GetObjectIdentifier(n *kyaml.RNode) string {
	return fmt.Sprintf("%s%s%s%s", n.GetApiVersion(), n.GetKind(), n.GetNamespace(), n.GetName())
}
