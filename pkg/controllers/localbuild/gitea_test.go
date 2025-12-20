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

const (
	testServerDelayDuration = 2 * time.Second
	testTimeoutDuration     = 100 * time.Millisecond
)

func TestGetGiteaToken(t *testing.T) {
	// Create a test server that delays response longer than the context timeout
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(testServerDelayDuration)
	}))
	defer ts.Close()

	// Use a context with a short timeout to test timeout behavior
	ctx, cancel := context.WithTimeout(context.Background(), testTimeoutDuration)
	defer cancel()

	_, err := util.GetGiteaToken(ctx, ts.URL, "", "")
	require.Error(t, err)
}
