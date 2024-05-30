package helpers

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateKubernetesYaml(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("could not get current working directory")
	}

	cases := map[string]struct {
		expectErr bool
		inputPath string
	}{
		"invalidPath": {expectErr: true, inputPath: fmt.Sprintf("%s/invalid/path", cwd)},
		"notAbs":      {expectErr: true, inputPath: fmt.Sprintf("invalid/path")},
		"valid":       {expectErr: false, inputPath: fmt.Sprintf("%s/test-data/valid.yaml", cwd)},
		"notYaml":     {expectErr: true, inputPath: fmt.Sprintf("%s/test-data/notyaml.yaml", cwd)},
		"notk8s":      {expectErr: true, inputPath: fmt.Sprintf("%s/test-data/notk8s.yaml", cwd)},
	}

	for k := range cases {
		cErr := ValidateKubernetesYamlFile(cases[k].inputPath)
		if cases[k].expectErr && cErr == nil {
			t.Fatalf("%s expected error but did not receive error", k)
		}
		if !cases[k].expectErr && cErr != nil {
			t.Fatalf("%s did not expect error but received error", k)
		}
	}
}

func TestParsePackageStrings(t *testing.T) {
	cases := map[string]struct {
		expectErr  bool
		inputPaths []string
		remote     int
		local      int
	}{
		"allLocal": {expectErr: false, inputPaths: []string{"test-data", "."}, remote: 0, local: 2},
		"allRemote": {expectErr: false, inputPaths: []string{
			"https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?timeout=120&ref=v3.3.1",
			"git@github.com:owner/repo//examples",
		}, remote: 2, local: 0},
		"mix": {expectErr: false, inputPaths: []string{
			"https://github.com/kubernetes-sigs/kustomize//examples/multibases/dev/?timeout=120&ref=v3.3.1",
			"test-data",
		}, remote: 1, local: 1},
		"invalidLocalPath": {expectErr: true, inputPaths: []string{
			"does-not-exist",
		}, remote: 0, local: 0},
		"invalidRemotePath": {expectErr: true, inputPaths: []string{
			"https://   github.com/kubernetes-sigs/kustomize//examples",
		}, remote: 0, local: 0},
	}

	for k := range cases {
		c := cases[k]
		remote, local, err := ParsePackageStrings(c.inputPaths)
		if cases[k].expectErr {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		assert.Equal(t, c.remote, len(remote))
		assert.Equal(t, c.local, len(local))
	}
}
