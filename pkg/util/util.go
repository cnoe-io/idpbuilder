package util

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	mathrand "math/rand"
	"path/filepath"
	"strings"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	chars           = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits          = "0123456789"
	specialChars    = `!#$%&'()*+,-./:;<=>?@[]^_{|}~`
	passwordLength  = 40
	numSpecialChars = 3
	numDigits       = 3
)

func GetCLIStartTimeAnnotationValue(annotations map[string]string) (string, error) {
	if annotations == nil {
		return "", fmt.Errorf("this object's annotation is nil")
	}
	timeStamp, ok := annotations[v1alpha1.CliStartTimeAnnotation]
	if ok {
		return timeStamp, nil
	}
	return "", fmt.Errorf("expected annotation, %s, not found", v1alpha1.CliStartTimeAnnotation)
}

func SetCLIStartTimeAnnotationValue(annotations map[string]string, timeStamp string) {
	if timeStamp != "" && annotations != nil {
		annotations[v1alpha1.CliStartTimeAnnotation] = timeStamp
	}
}

func SetLastObservedSyncTimeAnnotationValue(annotations map[string]string, timeStamp string) {
	if timeStamp != "" && annotations != nil {
		annotations[v1alpha1.LastObservedCLIStartTimeAnnotation] = timeStamp
	}
}

func GetLastObservedSyncTimeAnnotationValue(annotations map[string]string) (string, error) {
	if annotations == nil {
		return "", fmt.Errorf("this object's annotation is nil")
	}
	timeStamp, ok := annotations[v1alpha1.LastObservedCLIStartTimeAnnotation]
	if ok {
		return timeStamp, nil
	}
	return "", fmt.Errorf("expected annotation, %s, not found", v1alpha1.LastObservedCLIStartTimeAnnotation)
}

func UpdateSyncAnnotation(ctx context.Context, kubeClient client.Client, obj client.Object) error {
	timeStamp, err := GetCLIStartTimeAnnotationValue(obj.GetAnnotations())
	if err != nil {
		return err
	}
	annotations := make(map[string]string, 1)
	SetLastObservedSyncTimeAnnotationValue(annotations, timeStamp)
	// MUST be unstructured to avoid managing fields we do not care about.
	u := unstructured.Unstructured{}
	u.SetAnnotations(annotations)
	u.SetName(obj.GetName())
	u.SetNamespace(obj.GetNamespace())
	u.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())

	return kubeClient.Patch(ctx, &u, client.Apply, client.ForceOwnership, client.FieldOwner(v1alpha1.FieldManager))
}

func GeneratePassword() (string, error) {
	passChars := make([]string, passwordLength)
	validChars := fmt.Sprintf("%s%s%s", chars, digits, specialChars)

	for i := 0; i < numSpecialChars; i++ {
		c, err := getRandElement(specialChars)
		if err != nil {
			return "", err
		}
		passChars = append(passChars, c)
	}

	for i := 0; i < numDigits; i++ {
		c, err := getRandElement(digits)
		if err != nil {
			return "", err
		}
		passChars = append(passChars, c)
	}

	for i := 0; i < passwordLength-numDigits-numSpecialChars; i++ {
		c, err := getRandElement(validChars)
		if err != nil {
			return "", err
		}
		passChars = append(passChars, c)
	}

	seed, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return "", err
	}

	r := mathrand.New(mathrand.NewSource(seed.Int64()))
	r.Shuffle(len(passChars), func(i, j int) {
		passChars[i], passChars[j] = passChars[j], passChars[i]
	})

	return strings.Join(passChars, ""), nil
}

func getRandElement(input string) (string, error) {
	position, err := rand.Int(rand.Reader, big.NewInt(int64(len(input))))
	if err != nil {
		return "", err
	}

	return string(input[position.Int64()]), nil
}

func IsYamlFile(input string) bool {
	extension := filepath.Ext(input)
	return extension == ".yaml" || extension == ".yml"
}
