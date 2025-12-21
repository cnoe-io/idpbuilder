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
	"github.com/cnoe-io/idpbuilder/pkg/status"
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
	registryConfigUsage = "List of paths to mount as the registry config, uses the first one that exists"
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
	noExitUsage       = "When set, idpbuilder will not exit after all packages are synced. Useful for continuously syncing local directories."
	statusOutputUsage = "Status output mode. Supported values: auto (default), simple, verbose, none. 'auto' uses inline status in terminals, 'simple' shows status without inline updates, 'verbose' shows detailed logs, 'none' disables status output."
	noColorUsage      = "Disable colored output for both logs and status reporting."
	quietUsage        = "Suppress status output (equivalent to --status-output=none)."

	// Status output modes
	statusOutputAuto    = "auto"
	statusOutputSimple  = "simple"
	statusOutputVerbose = "verbose"
	statusOutputNone    = "none"
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
	registryConfig            []string
	packageCustomizationFiles []string
	noExit                    bool
	protocol                  string
	host                      string
	ingressHost               string
	port                      string
	pathRouting               bool
	statusOutput              string
	noColor                   bool
	quiet                     bool
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
	CreateCmd.PersistentFlags().StringVar(&kubeVersion, "kube-version", "v1.33.1", kubeVersionUsage)
	CreateCmd.PersistentFlags().StringVar(&extraPortsMapping, "extra-ports", "", extraPortsMappingUsage)
	CreateCmd.PersistentFlags().StringVar(&kindConfigPath, "kind-config", "", kindConfigPathUsage)
	CreateCmd.PersistentFlags().StringSliceVar(&registryConfig, "registry-config", []string{}, registryConfigUsage)
	CreateCmd.PersistentFlags().Lookup("registry-config").NoOptDefVal = "$XDG_RUNTIME_DIR/containers/auth.json,$HOME/.docker/config.json"

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

	// status and output related flags
	CreateCmd.Flags().StringVar(&statusOutput, "status-output", "auto", statusOutputUsage)
	CreateCmd.Flags().BoolVar(&noColor, "no-color", false, noColorUsage)
	CreateCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, quietUsage)
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

	var localFiles []string
	var localDirs []string
	var remotePaths []string

	if len(extraPackages) > 0 {
		r, f, d, pErr := helpers.ParsePackageStrings(extraPackages)
		if pErr != nil {
			return pErr
		}
		localFiles = f
		localDirs = d
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

	// If registry-config is unset we pass nil
	// If registry-config is change (--registry-config=foo) we pass the new value
	// If registry-config is set but unchanged (--registry-confg) we pass ""
	maybeRegistryConfig := []string{}
	if cmd.Flags().Changed("registry-config") {
		maybeRegistryConfig = registryConfig
	}

	// Determine if color should be enabled
	// Priority: --no-color flag > --color flag > default
	useColor := helpers.ColoredOutput
	if noColor {
		useColor = false
	}

	// Determine status output mode
	// Priority: --quiet flag > --status-output flag > default
	statusMode := statusOutput
	if quiet {
		statusMode = statusOutputNone
	}

	// Validate status output mode
	validModes := map[string]bool{
		statusOutputAuto:    true,
		statusOutputSimple:  true,
		statusOutputVerbose: true,
		statusOutputNone:    true,
	}
	if !validModes[statusMode] {
		return fmt.Errorf("invalid status-output value: %s. Supported values are: %s, %s, %s, %s",
			statusMode, statusOutputAuto, statusOutputSimple, statusOutputVerbose, statusOutputNone)
	}

	// Create status reporter based on mode
	var reporter *status.Reporter
	if statusMode != statusOutputNone && statusMode != statusOutputVerbose {
		reporter = status.NewReporter(useColor)
		if statusMode == statusOutputSimple {
			reporter.SetSimpleMode(true)
		}
		reporter.AddStep("cluster", "Creating Kubernetes cluster")
		reporter.AddStep("crds", "Installing Custom Resource Definitions")
		reporter.AddStep("networking", "Configuring networking and certificates")
		reporter.AddStep("resources", "Creating platform resources")
		reporter.AddStep("packages", "Installing and syncing packages")
	}

	opts := build.NewBuildOptions{
		Name:              buildName,
		KubeVersion:       kubeVersion,
		KubeConfigPath:    kubeConfigPath,
		KindConfigPath:    kindConfigPath,
		ExtraPortsMapping: extraPortsMapping,
		RegistryConfig:    maybeRegistryConfig,

		TemplateData: v1alpha1.BuildCustomizationSpec{
			Protocol:       protocol,
			Host:           host,
			IngressHost:    ingressHost,
			Port:           port,
			UsePathRouting: pathRouting,
			StaticPassword: devPassword,
		},

		CustomPackageFiles:   localFiles,
		CustomPackageDirs:    localDirs,
		CustomPackageUrls:    remotePaths,
		ExitOnSync:           exitOnSync,
		PackageCustomization: o,

		Scheme:         k8s.GetScheme(),
		CancelFunc:     ctxCancel,
		StatusReporter: reporter,
	}

	b := build.NewBuild(opts)

	// Ensure summary is always printed if reporter exists
	defer func() {
		if reporter != nil {
			reporter.Summary()
		}
	}()

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

	_, _, _, err = helpers.ParsePackageStrings(extraPackages)
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
