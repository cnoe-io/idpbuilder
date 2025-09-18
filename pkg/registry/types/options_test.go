package types

import "testing"

func TestConnectionOptions(t *testing.T) {
	opts := &ConnectionOptions{
		UserAgent: "test-client/1.0",
		Debug:     true,
	}
	if opts.UserAgent != "test-client/1.0" {
		t.Errorf("expected test-client/1.0")
	}
}