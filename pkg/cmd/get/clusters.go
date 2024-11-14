package get

import (
	"context"
	"fmt"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/helpers"
	"github.com/cnoe-io/idpbuilder/pkg/kind"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
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

	provider := cluster.NewProvider(cluster.ProviderWithLogger(kind.KindLoggerFromLogr(&logger)), detectOpt)
	clusters, err := provider.List()
	if err != nil {
		logger.Error(err, "failed to list clusters.")
	}

	for _, c := range clusters {
		fmt.Printf("Cluster: %s\n", c)
		var nodeList corev1.NodeList
		err := cli.List(context.TODO(), &nodeList)
		if err != nil {
			logger.Error(err, "failed to list nodes for cluster: %s", c)
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
