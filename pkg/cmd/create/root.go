package create

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/build"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/helpers"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	// Flags
	recreateCluster   bool
	buildName         string
	kubeVersion       string
	extraPortsMapping string
	kindConfigPath    string
	// TODO: Remove extraPackagesDirs after 0.6.0 release
	extraPackagesDirs         []string
	extraPackages             []string
	packageCustomizationFiles []string
	noExit                    bool
	protocol                  string
	host                      string
	ingressHost               string
	port                      string
	pathRouting               bool
)

var CreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "(Re)Create an IDP cluster",
	Long:    ``,
	RunE:    create,
	PreRunE: preCreateE,
}

func init() {
	// cluster related flags
	CreateCmd.PersistentFlags().BoolVar(&recreateCluster, "recreate", false, "Delete cluster first if it already exists.")
	CreateCmd.PersistentFlags().StringVar(&buildName, "build-name", "localdev", "Name for build (Prefix for kind cluster name, pod names, etc).")
	CreateCmd.PersistentFlags().StringVar(&kubeVersion, "kube-version", "v1.29.2", "Version of the kind kubernetes cluster to create.")
	CreateCmd.PersistentFlags().StringVar(&extraPortsMapping, "extra-ports", "", "List of extra ports to expose on the docker container and kubernetes cluster as nodePort (e.g. \"22:32222,9090:39090,etc\").")
	CreateCmd.PersistentFlags().StringVar(&kindConfigPath, "kind-config", "", "Path of the kind config file to be used instead of the default.")

	// in-cluster resources related flags
	CreateCmd.PersistentFlags().StringVar(&host, "host", "cnoe.localtest.me", "Host name to access resources in this cluster.")
	CreateCmd.PersistentFlags().StringVar(&ingressHost, "ingress-host-name", "", "Host name used by ingresses. Useful when you have another proxy in front of ingress-nginx that idpbuilder provisions.")
	CreateCmd.PersistentFlags().StringVar(&protocol, "protocol", "https", "Protocol to use to access web UIs. http or https.")
	CreateCmd.PersistentFlags().StringVar(&port, "port", "8443", "Port number under which idpBuilder tools are accessible.")
	CreateCmd.PersistentFlags().BoolVar(&pathRouting, "use-path-routing", false, "When set to true, web UIs are exposed under single domain name.")
	// TODO: Remove package-dir and deprecation notice after 0.6.0 release
	CreateCmd.Flags().StringSliceVar(&extraPackagesDirs, "package-dir", []string{}, "Paths to directories containing custom packages")
	CreateCmd.Flags().MarkDeprecated("package-dir", "use --package instead")
	CreateCmd.Flags().StringSliceVarP(&extraPackages, "package", "p", []string{}, "Paths to locations containing custom packages")
	CreateCmd.Flags().StringSliceVarP(&packageCustomizationFiles, "package-custom-file", "c", []string{}, "Name of the package and the path to file to customize the package with. e.g. argocd:/tmp/argocd.yaml")
	// idpbuilder related flags
	CreateCmd.Flags().BoolVarP(&noExit, "no-exit", "n", true, "When set, idpbuilder will not exit after all packages are synced. Useful for continuously syncing local directories.")
}

func preCreateE(cmd *cobra.Command, args []string) error {
	return helpers.SetLogger()
}

func create(cmd *cobra.Command, args []string) error {
	ctx, ctxCancel := context.WithCancel(ctrl.SetupSignalHandler())
	defer ctxCancel()

	kubeConfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	protocol = strings.ToLower(protocol)
	host = strings.ToLower(host)
	if ingressHost == "" {
		ingressHost = host
	}

	err := validate()
	if err != nil {
		return err
	}

	var absDirPaths []string
	var remotePaths []string

	// TODO: Remove this block after deprecation
	if len(extraPackagesDirs) > 0 {
		r, l, pErr := helpers.ParsePackageStrings(extraPackagesDirs)
		if pErr != nil {
			return pErr
		}
		absDirPaths = l
		remotePaths = r
	}

	if len(extraPackages) > 0 {
		r, l, pErr := helpers.ParsePackageStrings(extraPackages)
		if pErr != nil {
			return pErr
		}
		absDirPaths = l
		remotePaths = r
	}

	o := make(map[string]v1alpha1.PackageCustomization)
	for i := range packageCustomizationFiles {
		c, pErr := getPackageCustomFile(packageCustomizationFiles[i])
		if pErr != nil {
			return pErr
		}
		o[c.Name] = c
	}

	exitOnSync := true
	if cmd.Flags().Changed("no-exit") {
		exitOnSync = !noExit
	}

	opts := build.NewBuildOptions{
		Name:              buildName,
		KubeVersion:       kubeVersion,
		KubeConfigPath:    kubeConfigPath,
		KindConfigPath:    kindConfigPath,
		ExtraPortsMapping: extraPortsMapping,

		TemplateData: util.CorePackageTemplateConfig{
			Protocol:       protocol,
			Host:           host,
			IngressHost:    ingressHost,
			Port:           port,
			UsePathRouting: pathRouting,
		},

		CustomPackageDirs:    absDirPaths,
		CustomPackageUrls:    remotePaths,
		ExitOnSync:           exitOnSync,
		PackageCustomization: o,

		Scheme:     k8s.GetScheme(),
		CancelFunc: ctxCancel,
	}

	b := build.NewBuild(opts)

	if err := b.Run(ctx, recreateCluster); err != nil {
		return err
	}

	subDomain := "argocd."
	subPath := ""

	if pathRouting == true {
		subDomain = ""
		subPath = "argocd"
	}

	fmt.Print("\n\n########################### Finished Creating IDP Successfully! ############################\n\n\n")
	fmt.Printf("Can Access ArgoCD at %s\nUsername: admin\n", fmt.Sprintf("%s://%s%s:%s/%s", protocol, subDomain, host, port, subPath))
	fmt.Print(`Password can be retrieved by running: idpbuilder get secrets -p argocd`, "\n")

	return nil
}

func validate() error {
	if buildName == "" {
		return fmt.Errorf("must specify build-name")
	}

	_, err := url.Parse(fmt.Sprintf("%s://%s:%s", protocol, host, port))
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	for i := range packageCustomizationFiles {
		_, pErr := getPackageCustomFile(packageCustomizationFiles[i])
		if pErr != nil {
			return pErr
		}
	}

	_, _, err = helpers.ParsePackageStrings(extraPackagesDirs)
	return err
}

func getPackageCustomFile(input string) (v1alpha1.PackageCustomization, error) {
	// the format should be `<package-name>:<path-to-file>`
	s := strings.Split(input, ":")
	if len(s) != 2 {
		return v1alpha1.PackageCustomization{}, fmt.Errorf("ensure %s is formatted as <package-name>:<path-to-file>", input)
	}

	paths, err := helpers.GetAbsFilePaths([]string{s[1]}, false)
	if err != nil {
		return v1alpha1.PackageCustomization{}, err
	}

	err = helpers.ValidateKubernetesYamlFile(paths[0])
	if err != nil {
		return v1alpha1.PackageCustomization{}, err
	}

	corePkgs := map[string]struct{}{v1alpha1.ArgoCDPackageName: {}, v1alpha1.GiteaPackageName: {}, v1alpha1.IngressNginxPackageName: {}}
	name := s[0]
	_, ok := corePkgs[name]
	if !ok {
		return v1alpha1.PackageCustomization{}, fmt.Errorf("customization for %s not supported", name)
	}
	return v1alpha1.PackageCustomization{
		Name:     name,
		FilePath: paths[0],
	}, nil
}
