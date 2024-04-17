package helpers

import (
	"fmt"
	"os"
	"testing"
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
