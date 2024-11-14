package get

import (
	"context"
	"fmt"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/helpers"
	"github.com/cnoe-io/idpbuilder/pkg/kind"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/clientcmd/api"
	"os"
	"sigs.k8s.io/kind/pkg/cluster"
)

var ClustersCmd = &cobra.Command{
	Use:     "clusters",
	Short:   "Get idp clusters",
	Long:    ``,
	RunE:    list,
	PreRunE: preClustersE,
}

var kubeCfgPath string

func preClustersE(cmd *cobra.Command, args []string) error {
	return helpers.SetLogger()
}

// findClusterByName searches for a cluster by name in the kubeconfig
func findClusterByName(config *api.Config, name string) (*api.Cluster, bool) {
	cluster, exists := config.Clusters[name]
	return cluster, exists
}

func list(cmd *cobra.Command, args []string) error {
	logger := helpers.CmdLogger

	detectOpt, err := util.DetectKindNodeProvider()
	if err != nil {
		logger.Error(err, "failed to detect the provider.")
		os.Exit(1)
	}

	kubeConfig, err := helpers.GetKubeConfig()
	if err != nil {
		logger.Error(err, "failed to create the kube config.")
		os.Exit(1)
	}

	cli, err := helpers.GetKubeClient(kubeConfig)
	if err != nil {
		logger.Error(err, "failed to create the kube client.")
		os.Exit(1)
	}

	// List the idp builder clusters according to the provider: podman or docker
	provider := cluster.NewProvider(cluster.ProviderWithLogger(kind.KindLoggerFromLogr(&logger)), detectOpt)
	clusters, err := provider.List()
	if err != nil {
		logger.Error(err, "failed to list clusters.")
	}

	for _, cluster := range clusters {
		fmt.Printf("Cluster: %s\n", cluster)

		config, err := helpers.LoadKubeConfig()
		if err != nil {
			logger.Error(err, "failed to load the kube config.")
		}

		// Search about the idp cluster within the kubeconfig file
		cluster, found := findClusterByName(config, "kind-"+cluster)
		if !found {
			fmt.Printf("Cluster %q not found\n", cluster)
		} else {
			fmt.Printf("URL of the kube API server: %s\n", cluster.Server)
			fmt.Printf("TLS Verify: %t\n", cluster.InsecureSkipTLSVerify)
		}

		var nodeList corev1.NodeList
		err = cli.List(context.TODO(), &nodeList)
		if err != nil {
			logger.Error(err, "failed to list nodes for cluster: %s", cluster)
		}

		for _, node := range nodeList.Items {
			nodeName := node.Name
			fmt.Printf("  Node: %s\n", nodeName)

			for _, addr := range node.Status.Addresses {
				switch addr.Type {
				case corev1.NodeInternalIP:
					fmt.Printf("  Internal IP: %s\n", addr.Address)
				case corev1.NodeExternalIP:
					fmt.Printf("  External IP: %s\n", addr.Address)
				}
			}
			fmt.Println("----------")
		}
	}
	return nil
}
