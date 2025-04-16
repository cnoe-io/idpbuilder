package localbuild

import (
	"context"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetGiteaToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Second * 35)
	}))
	defer ts.Close()
	ctx := context.Background()
	_, err := util.GetGiteaToken(ctx, ts.URL, "", "")
	require.Error(t, err)
}
