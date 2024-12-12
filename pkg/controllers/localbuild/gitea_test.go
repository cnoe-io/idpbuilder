package localbuild

import (
	"context"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGiteaInternalBaseUrl(t *testing.T) {
	c := v1alpha1.BuildCustomizationSpec{
		Protocol:       "http",
		Port:           "8080",
		Host:           "cnoe.localtest.me",
		UsePathRouting: false,
	}

	s := giteaInternalBaseUrl(c)
	assert.Equal(t, "http://gitea.cnoe.localtest.me:8080", s)
	c.UsePathRouting = true
	s = giteaInternalBaseUrl(c)
	assert.Equal(t, "http://cnoe.localtest.me:8080/gitea", s)
}

func TestGetGiteaToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second * 35)
	}))
	defer ts.Close()
	ctx := context.Background()
	_, err := util.GetGiteaToken(ctx, ts.URL, "", "")
	require.Error(t, err)
}
