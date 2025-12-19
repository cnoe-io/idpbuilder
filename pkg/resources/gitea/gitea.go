package gitea

import (
	"context"
	"embed"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:embed resources/k8s/*
var installGiteaFS embed.FS

// GetInstallFS returns the embedded filesystem containing Gitea installation resources
func GetInstallFS() embed.FS {
	return installGiteaFS
}

// RawGiteaInstallResources returns raw Gitea installation manifests
func RawGiteaInstallResources(templateData any, config v1alpha1.PackageCustomization, scheme *runtime.Scheme) ([][]byte, error) {
	return k8s.BuildCustomizedManifests(config.FilePath, "resources/k8s", installGiteaFS, scheme, templateData)
}

// NewGiteaAdminSecret creates a new Gitea admin secret with the given password
func NewGiteaAdminSecret(password string) corev1.Secret {
	obj := util.GiteaAdminSecretObject()
	obj.StringData = map[string]string{
		"username": v1alpha1.GiteaAdminUserName,
		"password": password,
	}
	return obj
}

// SetGiteaToken creates or updates the Gitea admin token in the secret
func SetGiteaToken(ctx context.Context, kubeClient client.Client, secret corev1.Secret, baseUrl string) error {
	_, ok := secret.Data[util.GiteaAdminTokenFieldName]
	if ok {
		return nil
	}

	u := unstructured.Unstructured{}
	u.SetName(util.GiteaAdminSecret)
	u.SetNamespace(util.GiteaNamespace)
	u.SetGroupVersionKind(corev1.SchemeGroupVersion.WithKind("Secret"))

	user, ok := secret.Data["username"]
	if !ok {
		return fmt.Errorf("username field not found in gitea secret")
	}

	pass, ok := secret.Data["password"]
	if !ok {
		return fmt.Errorf("password field not found in gitea secret")
	}

	t, err := util.GetGiteaToken(ctx, baseUrl, string(user), string(pass))
	if err != nil {
		return fmt.Errorf("getting gitea token: %w", err)
	}

	token := base64.StdEncoding.EncodeToString([]byte(t))
	err = unstructured.SetNestedField(u.Object, token, "data", util.GiteaAdminTokenFieldName)
	if err != nil {
		return fmt.Errorf("setting gitea token field: %w", err)
	}

	return kubeClient.Patch(ctx, &u, client.Apply, client.ForceOwnership, client.FieldOwner(v1alpha1.FieldManager))
}

// CheckGiteaEndpoint checks if the Gitea API endpoint is ready
func CheckGiteaEndpoint(baseUrl string) (bool, error) {
	c := util.GetHttpClient()
	resp, err := c.Get(baseUrl)
	if err != nil {
		return false, err
	}
	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return false, nil
		}
	}
	return true, nil
}
