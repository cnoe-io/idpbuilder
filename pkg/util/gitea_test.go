package util

import (
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestGiteaBaseUrl(t *testing.T) {
	c := v1alpha1.BuildCustomizationSpec{
		Protocol:       "http",
		Port:           "8080",
		Host:           "cnoe.localtest.me",
		UsePathRouting: false,
	}

	s := GiteaBaseUrl(c)
	assert.Equal(t, "http://gitea.cnoe.localtest.me:8080", s)
	c.UsePathRouting = true
	s = GiteaBaseUrl(c)
	assert.Equal(t, "http://cnoe.localtest.me:8080/gitea", s)
}
