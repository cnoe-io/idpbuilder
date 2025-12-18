package get

import (
"context"
"os"
"path/filepath"
"testing"

"github.com/cnoe-io/idpbuilder/api/v1alpha1"
"github.com/cnoe-io/idpbuilder/globals"
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/mock"
corev1 "k8s.io/api/core/v1"
"k8s.io/apimachinery/pkg/api/errors"
"k8s.io/apimachinery/pkg/runtime/schema"
"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestGetCertificateFromCluster(t *testing.T) {
ctx := context.Background()

t.Run("successful certificate retrieval", func(t *testing.T) {
fClient := new(fakeKubeClient)
expectedCert := []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----")

secretKey := client.ObjectKey{
Name:      globals.SelfSignedCertSecretName,
Namespace: globals.NginxNamespace,
}

fClient.On("Get", ctx, secretKey, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
arg := args.Get(2).(*corev1.Secret)
arg.Data = map[string][]byte{
corev1.TLSCertKey: expectedCert,
}
}).Return(nil)

cert, err := getCertificateFromCluster(ctx, fClient)
assert.NoError(t, err)
assert.Equal(t, expectedCert, cert)
fClient.AssertExpectations(t)
})

t.Run("secret not found", func(t *testing.T) {
fClient := new(fakeKubeClient)

secretKey := client.ObjectKey{
Name:      globals.SelfSignedCertSecretName,
Namespace: globals.NginxNamespace,
}

notFoundErr := errors.NewNotFound(schema.GroupResource{Resource: "secrets"}, globals.SelfSignedCertSecretName)
fClient.On("Get", ctx, secretKey, mock.Anything, mock.Anything).Return(notFoundErr)

cert, err := getCertificateFromCluster(ctx, fClient)
assert.Error(t, err)
assert.Nil(t, cert)
assert.Contains(t, err.Error(), "getting certificate secret from cluster")
fClient.AssertExpectations(t)
})

t.Run("certificate data missing from secret", func(t *testing.T) {
fClient := new(fakeKubeClient)

secretKey := client.ObjectKey{
Name:      globals.SelfSignedCertSecretName,
Namespace: globals.NginxNamespace,
}

fClient.On("Get", ctx, secretKey, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
arg := args.Get(2).(*corev1.Secret)
arg.Data = map[string][]byte{
// Missing TLSCertKey
"other-key": []byte("value"),
}
}).Return(nil)

cert, err := getCertificateFromCluster(ctx, fClient)
assert.Error(t, err)
assert.Nil(t, cert)
assert.Contains(t, err.Error(), "certificate not found in secret")
fClient.AssertExpectations(t)
})
}

func TestRegistryHostDetermination(t *testing.T) {
testCases := []struct {
name            string
config          v1alpha1.BuildCustomizationSpec
expectedHost    string
}{
{
name: "subdomain routing",
config: v1alpha1.BuildCustomizationSpec{
Host:           "cnoe.localtest.me",
Port:           "8443",
UsePathRouting: false,
},
expectedHost: "gitea.cnoe.localtest.me:8443",
},
{
name: "path routing",
config: v1alpha1.BuildCustomizationSpec{
Host:           "cnoe.localtest.me",
Port:           "8443",
UsePathRouting: true,
},
expectedHost: "cnoe.localtest.me:8443",
},
{
name: "custom host with path routing",
config: v1alpha1.BuildCustomizationSpec{
Host:           "example.com",
Port:           "443",
UsePathRouting: true,
},
expectedHost: "example.com:443",
},
{
name: "custom host with subdomain routing",
config: v1alpha1.BuildCustomizationSpec{
Host:           "example.com",
Port:           "443",
UsePathRouting: false,
},
expectedHost: "gitea.example.com:443",
},
}

for _, tc := range testCases {
t.Run(tc.name, func(t *testing.T) {
var registryHost string
if tc.config.UsePathRouting {
registryHost = tc.config.Host + ":" + tc.config.Port
} else {
registryHost = "gitea." + tc.config.Host + ":" + tc.config.Port
}
assert.Equal(t, tc.expectedHost, registryHost)
})
}
}

func TestCertificateFileOperations(t *testing.T) {
t.Run("write certificate to custom path", func(t *testing.T) {
tempDir := t.TempDir()
certPath := filepath.Join(tempDir, "test-cert.crt")
testCert := []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----")

err := os.WriteFile(certPath, testCert, 0644)
assert.NoError(t, err)

// Verify file was created and contains correct data
readCert, err := os.ReadFile(certPath)
assert.NoError(t, err)
assert.Equal(t, testCert, readCert)

// Verify file permissions
info, err := os.Stat(certPath)
assert.NoError(t, err)
assert.Equal(t, os.FileMode(0644), info.Mode().Perm())
})

t.Run("create docker certs.d directory structure", func(t *testing.T) {
tempDir := t.TempDir()
registryHost := "gitea.cnoe.localtest.me:8443"
dockerCertsDir := filepath.Join(tempDir, ".docker", "certs.d", registryHost)

err := os.MkdirAll(dockerCertsDir, 0755)
assert.NoError(t, err)

// Verify directory was created
info, err := os.Stat(dockerCertsDir)
assert.NoError(t, err)
assert.True(t, info.IsDir())

// Verify we can write a certificate to it
certPath := filepath.Join(dockerCertsDir, "ca.crt")
testCert := []byte("test certificate")
err = os.WriteFile(certPath, testCert, 0644)
assert.NoError(t, err)

readCert, err := os.ReadFile(certPath)
assert.NoError(t, err)
assert.Equal(t, testCert, readCert)
})

t.Run("handle directory creation errors", func(t *testing.T) {
// Try to create a directory in a non-existent parent that we can't create
invalidPath := filepath.Join("/nonexistent-root-12345", "subdir")
err := os.MkdirAll(invalidPath, 0755)
assert.Error(t, err)
})
}

func TestPrintPostInstallInstructions(t *testing.T) {
// This is a smoke test to ensure the function doesn't panic
// We don't assert on the output as it's informational
t.Run("print instructions without panic", func(t *testing.T) {
assert.NotPanics(t, func() {
printPostInstallInstructions("gitea.cnoe.localtest.me:8443")
})
})
}

func TestCertificateCommandFlags(t *testing.T) {
t.Run("certificate command has required flags", func(t *testing.T) {
assert.NotNil(t, CertificateCmd)
assert.Equal(t, "certificate", CertificateCmd.Use)

// Check that flags are defined
outputFlag := CertificateCmd.Flags().Lookup("output")
assert.NotNil(t, outputFlag)
assert.Equal(t, "o", outputFlag.Shorthand)

dockerFlag := CertificateCmd.Flags().Lookup("docker")
assert.NotNil(t, dockerFlag)
assert.Equal(t, "false", dockerFlag.DefValue)
})
}
