package util

import (
	"strconv"
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var specialCharMap = make(map[string]struct{})

func TestGeneratePassword(t *testing.T) {
	for i := range specialChars {
		specialCharMap[string(specialChars[i])] = struct{}{}
	}

	for i := 0; i < 1000; i++ {
		p, err := GeneratePassword()
		if err != nil {
			t.Fatalf("error generating password: %v", err)
		}
		counts := make([]int, 3)
		for j := range p {
			counts[0] += 1
			c := string(p[j])
			_, ok := specialCharMap[c]
			if ok {
				counts[1] += 1
				continue
			}
			_, err := strconv.Atoi(c)
			if err == nil {
				counts[2] += 1
			}
		}
		if counts[0] != passwordLength {
			t.Fatalf("password length incorrect")
		}
		if counts[1] < numSpecialChars {
			t.Fatalf("min number of special chars not generated")
		}
		if counts[2] < numDigits {
			t.Fatalf("min number of digits not generated")
		}
	}
}

type MockObject struct {
	v1.ObjectMeta
}

func (m *MockObject) GetObjectKind() schema.ObjectKind {
	return nil
}

func (m *MockObject) DeepCopyObject() runtime.Object {
	return nil
}

func TestSetPackageLabels(t *testing.T) {
	testCases := []struct {
		name           string
		objectName     string
		initialLabels  map[string]string
		expectedLabels map[string]string
	}{
		{
			name:          "No initial labels",
			objectName:    "test-package",
			initialLabels: nil,
			expectedLabels: map[string]string{
				v1alpha1.PackageNameLabelKey: "test-package",
				v1alpha1.PackageTypeLabelKey: v1alpha1.PackageTypeLabelCustom,
			},
		},
		{
			name:       "With initial labels",
			objectName: "test-package-one",
			initialLabels: map[string]string{
				"existing":                   "label",
				v1alpha1.PackageNameLabelKey: "incorrect",
			},
			expectedLabels: map[string]string{
				"existing":                   "label",
				v1alpha1.PackageNameLabelKey: "test-package-one",
				v1alpha1.PackageTypeLabelKey: v1alpha1.PackageTypeLabelCustom,
			},
		},
		{
			name:          "ArgoCD package",
			objectName:    v1alpha1.ArgoCDPackageName,
			initialLabels: nil,
			expectedLabels: map[string]string{
				v1alpha1.PackageNameLabelKey: v1alpha1.ArgoCDPackageName,
				v1alpha1.PackageTypeLabelKey: v1alpha1.PackageTypeLabelCore,
			},
		},
		{
			name:          "Gitea package",
			objectName:    v1alpha1.GiteaPackageName,
			initialLabels: nil,
			expectedLabels: map[string]string{
				v1alpha1.PackageNameLabelKey: v1alpha1.GiteaPackageName,
				v1alpha1.PackageTypeLabelKey: v1alpha1.PackageTypeLabelCore,
			},
		},
		{
			name:          "IngressNginx package",
			objectName:    v1alpha1.IngressNginxPackageName,
			initialLabels: nil,
			expectedLabels: map[string]string{
				v1alpha1.PackageNameLabelKey: v1alpha1.IngressNginxPackageName,
				v1alpha1.PackageTypeLabelKey: v1alpha1.PackageTypeLabelCore,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			obj := &MockObject{
				ObjectMeta: v1.ObjectMeta{
					Name:   tc.objectName,
					Labels: tc.initialLabels,
				},
			}

			SetPackageLabels(obj)

			assert.Equal(t, tc.expectedLabels, obj.GetLabels())
		})
	}
}
