package build

import (
	"context"
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestIsCompatible(t *testing.T) {
	cfg := v1alpha1.BuildCustomizationSpec{
		Protocol:       "http",
		Host:           "cnoe.localtest.me",
		IngressHost:    "string",
		Port:           "8443",
		UsePathRouting: false,
		SelfSignedCert: "some-cert",
	}

	b := Build{
		name: "test",
		cfg:  cfg,
	}

	ctx := context.Background()
	fClient := new(fakeKubeClient)
	fClient.On("Get", ctx, client.ObjectKey{Name: "test"}, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*v1alpha1.Localbuild)
		arg.Spec.BuildCustomization = cfg
	}).Return(nil)

	ok, err := b.isCompatible(ctx, fClient)

	assert.NoError(t, err)
	fClient.AssertExpectations(t)
	require.True(t, ok)

	fClient = new(fakeKubeClient)
	fClient.On("Get", ctx, client.ObjectKey{Name: "test"}, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		arg := args.Get(2).(*v1alpha1.Localbuild)
		c := cfg
		c.Host = "not-right"
		arg.Spec.BuildCustomization = c
	}).Return(nil)

	ok, err = b.isCompatible(ctx, fClient)

	assert.Error(t, err)
	fClient.AssertExpectations(t)
	require.False(t, ok)

	fClient = new(fakeKubeClient)
	fClient.On("Get", ctx, client.ObjectKey{Name: "test"}, mock.Anything, mock.Anything).
		Return(k8serrors.NewNotFound(schema.GroupResource{}, "name"))

	ok, err = b.isCompatible(ctx, fClient)

	assert.NoError(t, err)
	fClient.AssertExpectations(t)
	require.True(t, ok)
}
