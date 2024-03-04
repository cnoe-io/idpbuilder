package create

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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
	extraPackagesDirs []string
	noExit            bool
	protocol          string
	host              string
	ingressHost       string
	port              string
	pathRouting       bool
)

var CreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "(Re)Create an IDP cluster",
	Long:    ``,
	RunE:    create,
	PreRunE: preCreateE,
}

func init() {
	CreateCmd.PersistentFlags().BoolVar(&recreateCluster, "recreate", false, "Delete cluster first if it already exists.")
	CreateCmd.PersistentFlags().StringVar(&buildName, "build-name", "localdev", "Name for build (Prefix for kind cluster name, pod names, etc).")
	CreateCmd.PersistentFlags().StringVar(&kubeVersion, "kube-version", "v1.27.3", "Version of the kind kubernetes cluster to create.")
	CreateCmd.PersistentFlags().StringVar(&extraPortsMapping, "extra-ports", "", "List of extra ports to expose on the docker container and kubernetes cluster as nodePort (e.g. \"22:32222,9090:39090,etc\").")
	CreateCmd.PersistentFlags().StringVar(&kindConfigPath, "kind-config", "", "Path of the kind config file to be used instead of the default.")
	CreateCmd.PersistentFlags().StringVar(&host, "host", "cnoe.localtest.me", "Host name to access resources in this cluster.")
	CreateCmd.PersistentFlags().StringVar(&ingressHost, "ingress-host-name", "", "Host name used by ingresses. Useful when you have another proxy infront of idpbuilder.")
	CreateCmd.PersistentFlags().StringVar(&protocol, "protocol", "https", "Protocol to use to access web UIs. http or https.")
	CreateCmd.PersistentFlags().StringVar(&port, "port", "8443", "Port number under which idpBuilder tools are accessible.")
	CreateCmd.PersistentFlags().BoolVar(&pathRouting, "use-path-routing", false, "When set to true, web UIs are exposed under single domain name.")
	CreateCmd.Flags().StringSliceVarP(&extraPackagesDirs, "package-dir", "p", []string{}, "Paths to custom packages")
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
	if len(extraPackagesDirs) > 0 {
		p, err := getPackageAbsDirs(extraPackagesDirs)
		if err != nil {
			return err
		}
		absDirPaths = p
	}

	exitOnSync := true
	if cmd.Flags().Changed("no-exit") {
		exitOnSync = !noExit
	}

	b := build.NewBuild(
		buildName, kubeVersion, kubeConfigPath, kindConfigPath, extraPortsMapping,
		util.CorePackageTemplateConfig{
			Protocol:       protocol,
			Host:           host,
			IngressHost:    ingressHost,
			Port:           port,
			UsePathRouting: pathRouting,
		},
		absDirPaths, exitOnSync, k8s.GetScheme(), ctxCancel,
	)

	if err := b.Run(ctx, recreateCluster); err != nil {
		return err
	}

	fmt.Print("\n\n########################### Finished Creating IDP Successfully! ############################\n\n\n")
	fmt.Printf("Can Access ArgoCD at https://argocd.cnoe.localtest.me:%s/\nUsername: admin\n", port)
	fmt.Print(`Password can be retrieved by running: kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d`, "\n")

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
	return nil
}

func getPackageAbsDirs(paths []string) ([]string, error) {
	out := make([]string, len(paths), len(paths))
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
		if !f.IsDir() {
			return nil, fmt.Errorf("given path is not a directory. %s", absPath)
		}
		out[i] = absPath
	}

	return out, nil
}
