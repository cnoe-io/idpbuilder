package k8s

import (
	"bytes"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	for _, objYaml := range bytes.Split(objYamls, []byte{'-', '-', '-'}) {
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
