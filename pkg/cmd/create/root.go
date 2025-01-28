package create

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/build"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/helpers"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

const (
	recreateClusterUsage   = "Delete cluster first if it already exists."
	buildNameUsage         = "Name for build (Prefix for kind cluster name, pod names, etc)."
	devPasswordUsage       = "Set the password \"developer\" for the admin user of the applications: argocd & gitea."
	kubeVersionUsage       = "Version of the kind kubernetes cluster to create."
	extraPortsMappingUsage = "List of extra ports to expose on the docker container and kubernetes cluster as nodePort " +
		"(e.g. \"22:32222,9090:39090,etc\")."
	kindConfigPathUsage = "Path or URL to the kind config file to be used instead of the default."
	hostUsage           = "Host name to access resources in this cluster."
	ingressHostUsage    = "Host name used by ingresses. Useful when you have another proxy in front of ingress-nginx that idpbuilder provisions."
	protocolUsage       = "Protocol to use to access web UIs. http or https."
	portUsage           = "Port number to use to access web UIs."
	pathRoutingUsage    = "When set to true, web UIs are exposed under single domain name. " +
		"e.g. \"https://cnoe.localtest.me/argocd\" instead of \"https://argocd.cnoe.localtest.me\""
	extraPackagesUsage             = "Paths to locations containing custom packages"
	packageCustomizationFilesUsage = "Name of the package and the path to file to customize the core packages with. " +
		"valid package names are: argocd, nginx, and gitea. e.g. argocd:/tmp/argocd.yaml"
	noExitUsage = "When set, idpbuilder will not exit after all packages are synced. Useful for continuously syncing local directories."
)

var (
	// Flags
	recreateCluster           bool
	buildName                 string
	devPassword               bool
	kubeVersion               string
	extraPortsMapping         string
	kindConfigPath            string
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
	Use:          "create",
	Short:        "(Re)Create an IDP cluster",
	Long:         ``,
	RunE:         create,
	PreRunE:      preCreateE,
	SilenceUsage: true,
}

func init() {
	// cluster related flags
	CreateCmd.PersistentFlags().BoolVar(&recreateCluster, "recreate", false, recreateClusterUsage)
	CreateCmd.PersistentFlags().StringVar(&buildName, "build-name", "localdev", buildNameUsage)
	CreateCmd.PersistentFlags().MarkDeprecated("build-name", "use --name instead.")
	CreateCmd.PersistentFlags().StringVar(&buildName, "name", "localdev", buildNameUsage)
	CreateCmd.PersistentFlags().BoolVar(&devPassword, "dev-password", false, devPasswordUsage)
	CreateCmd.PersistentFlags().StringVar(&kubeVersion, "kube-version", "v1.30.3", kubeVersionUsage)
	CreateCmd.PersistentFlags().StringVar(&extraPortsMapping, "extra-ports", "", extraPortsMappingUsage)
	CreateCmd.PersistentFlags().StringVar(&kindConfigPath, "kind-config", "", kindConfigPathUsage)

	// in-cluster resources related flags
	CreateCmd.PersistentFlags().StringVar(&host, "host", globals.DefaultHostName, hostUsage)
	CreateCmd.PersistentFlags().StringVar(&ingressHost, "ingress-host-name", "", ingressHostUsage)
	CreateCmd.PersistentFlags().StringVar(&protocol, "protocol", "https", protocolUsage)
	CreateCmd.PersistentFlags().StringVar(&port, "port", "8443", portUsage)
	CreateCmd.PersistentFlags().BoolVar(&pathRouting, "use-path-routing", false, pathRoutingUsage)
	CreateCmd.Flags().StringSliceVarP(&extraPackages, "package", "p", []string{}, extraPackagesUsage)
	CreateCmd.Flags().StringSliceVarP(&packageCustomizationFiles, "package-custom-file", "c", []string{}, packageCustomizationFilesUsage)
	// idpbuilder related flags
	CreateCmd.Flags().BoolVarP(&noExit, "no-exit", "n", true, noExitUsage)
}

func preCreateE(cmd *cobra.Command, args []string) error {
	return helpers.SetLogger()
}

func create(cmd *cobra.Command, args []string) error {

	ctx, ctxCancel := context.WithCancel(cmd.Context())
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

		TemplateData: v1alpha1.BuildCustomizationSpec{
			Protocol:       protocol,
			Host:           host,
			IngressHost:    ingressHost,
			Port:           port,
			UsePathRouting: pathRouting,
			StaticPassword: devPassword,
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

	if cmd.Context().Err() != nil {
		return context.Cause(cmd.Context())
	}

	printSuccessMsg()
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

	_, _, err = helpers.ParsePackageStrings(extraPackages)
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

func printSuccessMsg() {
	subDomain := "argocd."
	subPath := ""

	if pathRouting == true {
		subDomain = ""
		subPath = "argocd"
	}

	var argoURL string

	proxy := behindProxy()
	if proxy {
		argoURL = fmt.Sprintf("https://%s/argocd", host)
	} else {
		argoURL = fmt.Sprintf("%s://%s%s:%s/%s", protocol, subDomain, host, port, subPath)
	}

	fmt.Print("\n\n########################### Finished Creating IDP Successfully! ############################\n\n\n")
	fmt.Printf("Can Access ArgoCD at %s\nUsername: admin\n", argoURL)
	fmt.Print(`Password can be retrieved by running: idpbuilder get secrets -p argocd`, "\n")
}

func behindProxy() bool {
	// check if we are in codespaces: https://docs.github.com/en/codespaces/developing-in-a-codespace/default-environment-variables-for-your-codespace
	_, ok := os.LookupEnv("CODESPACES")
	return ok
}
