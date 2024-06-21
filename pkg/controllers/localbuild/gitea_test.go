package localbuild

import (
	"testing"

	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestGiteaInternalBaseUrl(t *testing.T) {
	c := util.CorePackageTemplateConfig{
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
