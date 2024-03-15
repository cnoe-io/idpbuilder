package helpers

import (
	"fmt"
	"os"
	"path/filepath"

	"sigs.k8s.io/kustomize/kyaml/kio"
)

func ValidateKubernetesYamlFile(absPath string) error {
	if !filepath.IsAbs(absPath) {
		return fmt.Errorf("given path is not an absolute path %s", absPath)
	}
	b, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed reading file: %s, err: %w", absPath, err)
	}
	n, err := kio.FromBytes(b)
	if err != nil {
		return fmt.Errorf("failed parsing file as kubernetes manifests file: %s, err: %w", absPath, err)
	}

	for i := range n {
		obj := n[i]
		if obj.IsNilOrEmpty() {
			return fmt.Errorf("given file %s contains an invalid kubenretes manifest", absPath)
		}
		if obj.GetKind() == "" || obj.GetApiVersion() == "" {
			return fmt.Errorf("given file %s contains an invalid kubenretes manifest", absPath)
		}
	}

	return nil
}

func GetAbsFilePaths(paths []string, isDir bool) ([]string, error) {
	out := make([]string, len(paths))
	for i := range paths {
		path := paths[i]
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("failed to validate path %s : %w", path, err)
		}
		f, err := os.Stat(absPath)
		if err != nil {
			return nil, fmt.Errorf("failed to validate path %s : %w", absPath, err)
		}
		if isDir && !f.IsDir() {
			return nil, fmt.Errorf("given path is not a directory. %s", absPath)
		}
		if !isDir && !f.Mode().IsRegular() {
			return nil, fmt.Errorf("give path is not a file. %s", absPath)
		}

		out[i] = absPath
	}

	return out, nil
}
