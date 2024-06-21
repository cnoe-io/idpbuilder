package build

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/stretchr/testify/mock"
	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type fakeKubeClient struct {
	mock.Mock
	client.Client
}

func (f *fakeKubeClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	args := f.Called(ctx, key, obj, opts)
	return args.Error(0)
}

func (f *fakeKubeClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	args := f.Called(ctx, obj, opts)
	return args.Error(0)
}

func TestCreateSelfSignedCertificate(t *testing.T) {
	sans := []string{"cnoe.io", "*.cnoe.io"}
	c, k, err := createSelfSignedCertificate(sans)
	assert.NilError(t, err)
	_, err = tls.X509KeyPair(c, k)
	assert.NilError(t, err)

	block, _ := pem.Decode(c)
	assert.Equal(t, "CERTIFICATE", block.Type)
	cert, err := x509.ParseCertificate(block.Bytes)
	assert.NilError(t, err)

	assert.Equal(t, 2, len(cert.DNSNames))
	expected := map[string]struct{}{
		"cnoe.io":   {},
		"*.cnoe.io": {},
	}

	for _, s := range cert.DNSNames {
		_, ok := expected[s]
		if ok {
			delete(expected, s)
		} else {
			t.Fatalf("unexpected key %s found", s)
		}
	}
	assert.Equal(t, 0, len(expected))
}

func TestGetOrCreateIngressCertificateAndKey(t *testing.T) {
	ctx := context.Background()
	fClient := new(fakeKubeClient)
	fClient.On("Get", ctx, client.ObjectKey{Name: globals.SelfSignedCertSecretName, Namespace: globals.NginxNamespace}, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*corev1.Secret)
		d := map[string][]byte{
			corev1.TLSPrivateKeyKey: []byte("abc"),
			corev1.TLSCertKey:       []byte("abc"),
		}
		arg.Data = d
	}).Return(nil)

	_, _, err := getOrCreateIngressCertificateAndKey(ctx, fClient, globals.SelfSignedCertSecretName, globals.NginxNamespace, []string{globals.DefaultHostName, globals.DefaultSANWildcard})
	assert.NilError(t, err)
	fClient.AssertExpectations(t)

	fClient = new(fakeKubeClient)
	fClient.On("Get", ctx, client.ObjectKey{Name: globals.SelfSignedCertSecretName, Namespace: globals.NginxNamespace}, mock.Anything, mock.Anything).
		Return(k8serrors.NewNotFound(schema.GroupResource{}, "name"))
	fClient.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)

	c, k, err := getOrCreateIngressCertificateAndKey(ctx, fClient, globals.SelfSignedCertSecretName, globals.NginxNamespace, []string{globals.DefaultHostName, globals.DefaultSANWildcard})
	assert.NilError(t, err)
	_, err = tls.X509KeyPair(c, k)
	assert.NilError(t, err)
}
