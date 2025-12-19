package gitea

import (
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestRawGiteaInstallResources(t *testing.T) {
	scheme := runtime.NewScheme()
	templateData := v1alpha1.BuildCustomizationSpec{
		Protocol:       "http",
		Host:           "cnoe.localtest.me",
		Port:           "8443",
		UsePathRouting: false,
	}
	config := v1alpha1.PackageCustomization{}

	resources, err := RawGiteaInstallResources(templateData, config, scheme)
	require.NoError(t, err)
	assert.NotEmpty(t, resources, "Expected at least one Gitea resource")
}

func TestNewGiteaAdminSecret(t *testing.T) {
	password := "test-password"
	secret := NewGiteaAdminSecret(password)

	assert.Equal(t, "gitea-credential", secret.Name)
	assert.Equal(t, "gitea", secret.Namespace)
	assert.Equal(t, "giteaAdmin", secret.StringData["username"])
	assert.Equal(t, password, secret.StringData["password"])
}
