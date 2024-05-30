package helpers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cnoe-io/idpbuilder/pkg/util"
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
			return fmt.Errorf("given file %s contains an invalid kubernetes manifest", absPath)
		}
		if obj.GetKind() == "" || obj.GetApiVersion() == "" {
			return fmt.Errorf("given file %s contains an invalid kubernetes manifest", absPath)
		}
	}

	return nil
}

func ParsePackageStrings(pkgStrings []string) ([]string, []string, error) {
	remote, local := make([]string, 0, 2), make([]string, 0, 2)
	for i := range pkgStrings {
		loc := pkgStrings[i]
		_, err := util.NewKustomizeRemote(loc)
		if err == nil {
			remote = append(remote, loc)
			continue
		}

		absPath, err := getAbsPath(loc, true)
		if err == nil {
			local = append(local, absPath)
			continue
		}
		return nil, nil, err
	}

	return remote, local, nil
}

func getAbsPath(path string, isDir bool) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("failed to validate path %s : %w", path, err)
	}
	f, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to validate path %s : %w", absPath, err)
	}
	if isDir && !f.IsDir() {
		return "", fmt.Errorf("given path is not a directory. %s", absPath)
	}
	if !isDir && !f.Mode().IsRegular() {
		return "", fmt.Errorf("give path is not a file. %s", absPath)
	}
	return absPath, nil
}

func GetAbsFilePaths(paths []string, isDir bool) ([]string, error) {
	out := make([]string, len(paths))
	for i := range paths {
		absPath, err := getAbsPath(paths[i], isDir)
		if err != nil {
			return nil, err
		}
		out[i] = absPath
	}
	return out, nil
}
