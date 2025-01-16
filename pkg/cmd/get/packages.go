package get

import (
	"context"
	"fmt"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/build"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/cnoe-io/idpbuilder/pkg/printer"
	"github.com/cnoe-io/idpbuilder/pkg/printer/types"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/spf13/cobra"
	"io"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

var PackagesCmd = &cobra.Command{
	Use:          "packages",
	Short:        "retrieve packages from the cluster",
	Long:         ``,
	RunE:         getPackagesE,
	SilenceUsage: true,
}

func getPackagesE(cmd *cobra.Command, args []string) error {
	ctx, ctxCancel := context.WithCancel(cmd.Context())
	defer ctxCancel()
	kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	opts := build.NewBuildOptions{
		KubeConfigPath: kubeConfigPath,
		Scheme:         k8s.GetScheme(),
		CancelFunc:     ctxCancel,
		TemplateData:   v1alpha1.BuildCustomizationSpec{},
	}

	b := build.NewBuild(opts)

	kubeConfig, err := b.GetKubeConfig()
	if err != nil {
		return fmt.Errorf("getting kube config: %w", err)
	}

	kubeClient, err := b.GetKubeClient(kubeConfig)
	if err != nil {
		return fmt.Errorf("getting kube client: %w", err)
	}

	return printPackages(ctx, os.Stdout, kubeClient, outputFormat)
}

// Print all the custom packages or based on package arguments passed using flag: -p
func printPackages(ctx context.Context, outWriter io.Writer, kubeClient client.Client, format string) error {
	packageList := []types.Package{}
	customPackages := v1alpha1.CustomPackageList{}
	var err error

	idpbuilderNamespace, err := getIDPNamespace(ctx, kubeClient)
	if err != nil {
		return fmt.Errorf("getting namespace: %w", err)
	}

	argocdBaseUrl, err := util.ArgocdBaseUrl(ctx)
	if err != nil {
		return fmt.Errorf("Error creating argocd Url: %v\n", err)
	}

	if len(packages) == 0 {
		// Get all custom packages
		customPackages, err = getPackages(ctx, kubeClient, idpbuilderNamespace)
		if err != nil {
			return fmt.Errorf("listing custom packages: %w", err)
		}
	} else {
		// Get the custom package using its name
		for _, name := range packages {
			cp, err := getPackageByName(ctx, kubeClient, idpbuilderNamespace, name)
			if err != nil {
				return fmt.Errorf("getting custom package %s: %w", name, err)
			}
			customPackages.Items = append(customPackages.Items, cp)
		}
	}

	for _, cp := range customPackages.Items {
		newPackage := types.Package{}
		newPackage.Name = cp.Name
		newPackage.Namespace = cp.Namespace
		newPackage.ArgocdRepository = argocdBaseUrl + "/applications/" + cp.Spec.ArgoCD.Namespace + "/" + cp.Spec.ArgoCD.Name
		// There is a GitRepositoryRefs when the project has been cloned to the internal git repository
		if cp.Status.GitRepositoryRefs != nil {
			newPackage.GitRepository = cp.Spec.InternalGitServeURL + "/" + v1alpha1.GiteaAdminUserName + "/" + idpbuilderNamespace + "-" + cp.Status.GitRepositoryRefs[0].Name
		} else {
			// Default branch reference
			ref := "main"
			if cp.Spec.RemoteRepository.Ref != "" {
				ref = cp.Spec.RemoteRepository.Ref
			}
			newPackage.GitRepository = cp.Spec.RemoteRepository.Url + "/tree/" + ref + "/" + cp.Spec.RemoteRepository.Path
		}

		newPackage.Status = strconv.FormatBool(cp.Status.Synced)

		packageList = append(packageList, newPackage)
	}

	packagePrinter := printer.PackagePrinter{
		Packages:  packageList,
		OutWriter: outWriter,
	}
	return packagePrinter.PrintOutput(format)
}

func getPackageByName(ctx context.Context, kubeClient client.Client, ns, name string) (v1alpha1.CustomPackage, error) {
	p := v1alpha1.CustomPackage{}
	return p, kubeClient.Get(ctx, client.ObjectKey{Name: name, Namespace: ns}, &p)
}

func getIDPNamespace(ctx context.Context, kubeClient client.Client) (string, error) {
	build, err := getLocalBuild(ctx, kubeClient)
	if err != nil {
		return "", err
	}
	// TODO: We assume that only one LocalBuild has been created for one cluster !
	idpNamespace := v1alpha1.FieldManager + "-" + build.Items[0].Name
	return idpNamespace, nil
}

func getLocalBuild(ctx context.Context, kubeClient client.Client) (v1alpha1.LocalbuildList, error) {
	localBuildList := v1alpha1.LocalbuildList{}
	return localBuildList, kubeClient.List(ctx, &localBuildList)
}

func getPackages(ctx context.Context, kubeClient client.Client, ns string) (v1alpha1.CustomPackageList, error) {
	packageList := v1alpha1.CustomPackageList{}
	return packageList, kubeClient.List(ctx, &packageList, client.InNamespace(ns))
}
