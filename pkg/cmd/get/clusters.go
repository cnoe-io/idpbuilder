package get

import (
	"context"
	"fmt"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/helpers"
	"github.com/cnoe-io/idpbuilder/pkg/kind"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/tools/clientcmd/api"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kind/pkg/cluster"
	"strings"
)

var ClustersCmd = &cobra.Command{
	Use:     "clusters",
	Short:   "Get idp clusters",
	Long:    ``,
	RunE:    list,
	PreRunE: preClustersE,
}

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

		// Search about the idp cluster within the kubeconfig file and show information
		c, found := findClusterByName(config, "kind-"+cluster)
		if !found {
			fmt.Printf("Cluster not found: %s\n", cluster)
		} else {
			fmt.Printf("URL of the kube API server: %s\n", c.Server)
			fmt.Printf("TLS Verify: %t\n", c.InsecureSkipTLSVerify)
		}
		fmt.Println("----------------------------------------")
	}

	// Let's check what the current node reports
	var nodeList corev1.NodeList
	err = cli.List(context.TODO(), &nodeList)
	if err != nil {
		logger.Error(err, "failed to list nodes for the current kube cluster.")
	}

	for _, node := range nodeList.Items {
		nodeName := node.Name
		fmt.Printf("Node: %s\n", nodeName)

		for _, addr := range node.Status.Addresses {
			switch addr.Type {
			case corev1.NodeInternalIP:
				fmt.Printf("Internal IP: %s\n", addr.Address)
			case corev1.NodeExternalIP:
				fmt.Printf("External IP: %s\n", addr.Address)
			}
		}

		// Show node capacity
		fmt.Printf("Capacity of the node: \n")
		printFormattedResourceList(node.Status.Capacity)

		// Show node allocated resources
		err = printAllocatedResources(context.Background(), cli, node.Name)
		if err != nil {
			logger.Error(err, "Failed to get the node's allocated resources.")
		}
		fmt.Println("--------------------")
	}

	return nil
}

func printFormattedResourceList(resources corev1.ResourceList) {
	// Define the fixed width for the resource name column (adjust as needed)
	nameWidth := 20

	for name, quantity := range resources {
		if strings.HasPrefix(string(name), "hugepages-") {
			continue
		}

		if name == corev1.ResourceMemory {
			// Convert memory from bytes to gigabytes (GB)
			memoryInBytes := quantity.Value()                           // .Value() gives the value in bytes
			memoryInGB := float64(memoryInBytes) / (1024 * 1024 * 1024) // Convert to GB
			fmt.Printf("  %-*s %.2f GB\n", nameWidth, name, memoryInGB)
		} else {
			// Format each line with the fixed name width and quantity
			fmt.Printf("  %-*s %s\n", nameWidth, name, quantity.String())
		}
	}
}

func printAllocatedResources(ctx context.Context, k8sClient client.Client, nodeName string) error {
	// List all pods on the specified node
	var podList corev1.PodList
	if err := k8sClient.List(ctx, &podList, client.MatchingFields{"spec.nodeName": nodeName}); err != nil {
		return fmt.Errorf("failed to list pods on node %s: %w", nodeName, err)
	}

	// Initialize counters for CPU and memory requests
	totalCPU := resource.NewQuantity(0, resource.DecimalSI)
	totalMemory := resource.NewQuantity(0, resource.BinarySI)

	// Sum up CPU and memory requests from each container in each pod
	for _, pod := range podList.Items {
		for _, container := range pod.Spec.Containers {
			if reqCPU, found := container.Resources.Requests[corev1.ResourceCPU]; found {
				totalCPU.Add(reqCPU)
			}
			if reqMemory, found := container.Resources.Requests[corev1.ResourceMemory]; found {
				totalMemory.Add(reqMemory)
			}
		}
	}

	// Display the total allocated resources
	fmt.Printf("Allocated resources on node:\n")
	fmt.Printf("  CPU Requests: %s\n", totalCPU.String())
	fmt.Printf("  Memory Requests: %s\n", totalMemory.String())

	return nil
}
