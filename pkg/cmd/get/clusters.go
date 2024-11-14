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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kind/pkg/cluster"
	"slices"
	"strings"
)

// ClusterManager holds the clients for the different idpbuilder clusters
type ClusterManager struct {
	clients map[string]client.Client // map of cluster name to client
}

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

	// TODO: Check if we need it or not like also if the new code handle the kubeconfig path passed as parameter
	_, err = helpers.GetKubeClient(kubeConfig)
	if err != nil {
		logger.Error(err, "failed to create the kube client.")
		os.Exit(1)
	}

	config, err := helpers.LoadKubeConfig()
	if err != nil {
		logger.Error(err, "failed to load the kube config.")
		os.Exit(1)
	}

	// List the idp builder clusters according to the provider: podman or docker
	provider := cluster.NewProvider(cluster.ProviderWithLogger(kind.KindLoggerFromLogr(&logger)), detectOpt)
	clusters, err := provider.List()
	if err != nil {
		logger.Error(err, "failed to list clusters.")
	}

	// Populate a list of Kube client for each cluster/context matching a idpbuilder cluster
	manager, _ := CreateKubeClientForEachIDPCluster(config, clusters)

	fmt.Printf("\n")
	for _, cluster := range clusters {
		fmt.Printf("Cluster: %s\n", cluster)

		// Search about the idp cluster within the kubeconfig file and show information
		c, found := findClusterByName(config, "kind-"+cluster)
		if !found {
			fmt.Printf("Cluster not found: %s\n", cluster)
		} else {
			fmt.Printf("URL of the kube API server: %s\n", c.Server)
			fmt.Printf("TLS Verify: %t\n", c.InsecureSkipTLSVerify)

			cli, err := GetClientForCluster(manager, cluster)
			if err != nil {
				logger.Error(err, "failed to get the cluster/context for the cluster: %s.", cluster)
			}
			// Print the external port that users can access using the ingress nginx proxy
			service := corev1.Service{}
			namespacedName := types.NamespacedName{
				Name:      "ingress-nginx-controller",
				Namespace: "ingress-nginx",
			}
			err = cli.Get(context.TODO(), namespacedName, &service)
			if err != nil {
				logger.Error(err, "failed to get the ingress service on the cluster.")
			}
			fmt.Printf("External Port: %d", findExternalHTTPSPort(service))

			// Let's check what the current node reports
			var nodeList corev1.NodeList
			err = cli.List(context.TODO(), &nodeList)
			if err != nil {
				logger.Error(err, "failed to list nodes for the current kube cluster.")
			}

			for _, node := range nodeList.Items {
				nodeName := node.Name
				fmt.Printf("\n\n")
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
			}
		}
		fmt.Println("----------------------------------------")
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

func findExternalHTTPSPort(service corev1.Service) int32 {
	var targetPort corev1.ServicePort
	for _, port := range service.Spec.Ports {
		if port.Name != "" && strings.HasPrefix(port.Name, "https-") {
			targetPort = port
			break
		}
	}
	return targetPort.Port
}

// GetClientForCluster returns the client for the specified cluster/context name
func GetClientForCluster(m *ClusterManager, clusterName string) (client.Client, error) {
	cl, exists := m.clients["kind-"+clusterName]
	if !exists {
		return nil, fmt.Errorf("no client found for cluster %q", clusterName)
	}
	return cl, nil
}

func CreateKubeClientForEachIDPCluster(config *api.Config, clusterList []string) (*ClusterManager, error) {
	// Initialize the ClusterManager with a map of kube Client
	manager := &ClusterManager{
		clients: make(map[string]client.Client),
	}

	for contextName := range config.Contexts {
		// Check if the kubconfig contains the cluster name
		// We remove the prefix "kind-" to find the cluster name from the slice
		if slices.Contains(clusterList, contextName[5:]) {
			cfg, err := clientcmd.NewNonInteractiveClientConfig(*config, contextName, &clientcmd.ConfigOverrides{}, nil).ClientConfig()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to build client for context %q: %v\n", contextName, err)
				continue
			}

			cl, err := client.New(cfg, client.Options{})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create client for context %q: %v\n", contextName, err)
				continue
			}

			manager.clients[contextName] = cl
			// fmt.Printf("Client created for context %q\n", contextName)
		}

	}
	return manager, nil
}
