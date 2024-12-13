package util

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"math"
	"math/big"
	mathrand "math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kind/pkg/cluster"
)

const (
	chars           = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digits          = "0123456789"
	specialChars    = `!#$%&'()*+,-./:;<=>?@[]^_{|}~`
	passwordLength  = 40
	numSpecialChars = 3
	numDigits       = 3
	StaticPassword  = "developer"
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

	return ApplyAnnotation(ctx, kubeClient, obj, annotations, client.ForceOwnership, client.FieldOwner(v1alpha1.FieldManager))
}

func ApplyAnnotation(ctx context.Context, kubeClient client.Client, obj client.Object, annotations map[string]string, opts ...client.PatchOption) error {
	// MUST be unstructured to avoid managing fields we do not care about.
	u := unstructured.Unstructured{}
	u.SetAnnotations(annotations)
	u.SetName(obj.GetName())
	u.SetNamespace(obj.GetNamespace())
	u.SetGroupVersionKind(obj.GetObjectKind().GroupVersionKind())
	return kubeClient.Patch(ctx, &u, client.Apply, opts...)
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

func GetHttpClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second, // from http.DefaultTransport
		}).DialContext,
	}
	return &http.Client{Transport: tr, Timeout: 30 * time.Second}
}

// DetectKindNodeProvider follows the kind CLI convention where:
// 1. if KIND_EXPERIMENTAL_PROVIDER env var is specified, it uses the value:
// 2. if env var is not specified, use the first available supported engine.
// https://github.com/kubernetes-sigs/kind/blob/ac81e7b64e06670132dae3486e64e531953ad58c/pkg/cluster/provider.go#L100-L114
func DetectKindNodeProvider() (cluster.ProviderOption, error) {
	switch p := os.Getenv("KIND_EXPERIMENTAL_PROVIDER"); p {
	case "podman":
		return cluster.ProviderWithPodman(), nil
	case "docker":
		return cluster.ProviderWithDocker(), nil
	case "nerdctl", "finch", "nerdctl.lima":
		return cluster.ProviderWithNerdctl(p), nil
	default:
		return cluster.DetectNodeProvider()
	}
}
