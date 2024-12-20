package get

import (
	"context"
	"fmt"
	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
	"github.com/cnoe-io/idpbuilder/pkg/cmd/helpers"
	"github.com/cnoe-io/idpbuilder/pkg/k8s"
	"github.com/cnoe-io/idpbuilder/pkg/kind"
	"github.com/cnoe-io/idpbuilder/pkg/printer"
	"github.com/cnoe-io/idpbuilder/pkg/util"
	"github.com/spf13/cobra"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type Cluster struct {
	Name         string
	URLKubeApi   string
	KubePort     int32
	TlsCheck     bool
	ExternalPort int32
	Nodes        []Node
}

type Node struct {
	Name       string
	InternalIP string
	ExternalIP string
	Capacity   Capacity
	Allocated  Allocated
}

type Capacity struct {
	Memory float64
	Pods   int64
	Cpu    int64
}

type Allocated struct {
	Cpu    string
	Memory string
}

var ClustersCmd = &cobra.Command{
	Use:          "clusters",
	Short:        "Get idp clusters",
	Long:         ``,
	RunE:         list,
	PreRunE:      preClustersE,
	SilenceUsage: true,
}

func preClustersE(cmd *cobra.Command, args []string) error {
	return helpers.SetLogger()
}

func list(cmd *cobra.Command, args []string) error {
	clusters, err := populateClusterList()
	if err != nil {
		return err
	} else {
		// Convert the list of the clusters to a Table of clusters and print the table using the format selected
		err := printClustersOutput(os.Stdout, clusters, outputFormat)
		if err != nil {
			return err
		} else {
			return nil
		}
	}
}

func printClustersOutput(outWriter io.Writer, clusters []Cluster, format string) error {
	switch format {
	case "json":
		return printer.PrintDataAsJson(clusters, outWriter)
	case "yaml":
		return printer.PrintDataAsYaml(clusters, outWriter)
	case "table":
		return printer.PrintTable(generateClusterTable(clusters), outWriter)
	default:

		return fmt.Errorf("output format %s is not supported", format)
	}
}

func populateClusterList() ([]Cluster, error) {
	logger := helpers.CmdLogger

	detectOpt, err := util.DetectKindNodeProvider()
	if err != nil {
		return nil, err
	}

	kubeConfig, err := helpers.GetKubeConfig()
	if err != nil {
		return nil, err
	}

	// TODO: Check if we need it or not like also if the new code handle the kubeconfig path passed as parameter
	_, err = helpers.GetKubeClient(kubeConfig)
	if err != nil {
		return nil, err
	}

	config, err := helpers.LoadKubeConfig()
	if err != nil {
		//logger.Error(err, "failed to load the kube config.")
		return nil, err
	}

	// Create an empty array of clusters to collect the information
	clusterList := []Cluster{}

	// List the idp builder clusters according to the provider: podman or docker
	provider := cluster.NewProvider(cluster.ProviderWithLogger(kind.KindLoggerFromLogr(&logger)), detectOpt)
	clusters, err := provider.List()
	if err != nil {
		return nil, err
	}

	// Populate a list of Kube client for each cluster/context matching an idpbuilder cluster
	manager, err := CreateKubeClientForEachIDPCluster(config, clusters)
	if err != nil {
		return nil, err
	}

	for _, cluster := range clusters {
		aCluster := Cluster{Name: cluster}

		// Search about the idp cluster within the kubeconfig file and show information
		c, found := findClusterByName(config, "kind-"+cluster)
		if !found {
			logger.Info(fmt.Sprintf("Cluster not found: %s within kube config file\n", cluster))
		} else {
			cli, err := GetClientForCluster(manager, cluster)
			if err != nil {
				return nil, err
			}
			logger.V(1).Info(fmt.Sprintf("Got the context for the cluster: %s.", cluster))

			// Print the external port mounted on the container and available also as ingress host port
			targetPort, err := findExternalHTTPSPort(cli, cluster)
			if err != nil {
				return nil, err
			} else {
				aCluster.ExternalPort = targetPort
			}

			aCluster.URLKubeApi = c.Server
			aCluster.TlsCheck = c.InsecureSkipTLSVerify

			// Print the internal port running the Kube API service
			kubeApiPort, err := findInternalKubeApiPort(cli)
			if err != nil {
				return nil, err
			} else {
				aCluster.KubePort = kubeApiPort
			}

			// Let's check what the current node reports
			var nodeList corev1.NodeList
			err = cli.List(context.TODO(), &nodeList)
			if err != nil {
				return nil, err
			}

			for _, node := range nodeList.Items {
				nodeName := node.Name

				aNode := Node{}
				aNode.Name = nodeName

				for _, addr := range node.Status.Addresses {
					switch addr.Type {
					case corev1.NodeInternalIP:
						aNode.InternalIP = addr.Address
					case corev1.NodeExternalIP:
						aNode.ExternalIP = addr.Address
					}
				}

				// Get Node capacity
				resources := node.Status.Capacity

				memory := resources[corev1.ResourceMemory]
				cpu := resources[corev1.ResourceCPU]
				pods := resources[corev1.ResourcePods]

				aNode.Capacity = Capacity{
					Memory: float64(memory.Value()) / (1024 * 1024 * 1024),
					Cpu:    cpu.Value(),
					Pods:   pods.Value(),
				}

				// Get Node Allocated resources
				allocated, err := printAllocatedResources(context.Background(), cli, node.Name)
				if err != nil {
					return nil, err
				}
				aNode.Allocated = allocated

				aCluster.Nodes = append(aCluster.Nodes, aNode)
			}

		}
		clusterList = append(clusterList, aCluster)
	}

	return clusterList, nil
}

func generateClusterTable(clusterTable []Cluster) metav1.Table {
	table := &metav1.Table{}
	table.ColumnDefinitions = []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string"},
		{Name: "External-Port", Type: "string"},
		{Name: "Kube-Api", Type: "string"},
		{Name: "TLS", Type: "string"},
		{Name: "Kube-Port", Type: "string"},
		{Name: "Nodes", Type: "string"},
	}
	for _, cluster := range clusterTable {
		row := metav1.TableRow{
			Cells: []interface{}{
				cluster.Name,
				cluster.ExternalPort,
				cluster.URLKubeApi,
				cluster.TlsCheck,
				cluster.KubePort,
				generateNodeData(cluster.Nodes),
			},
		}
		table.Rows = append(table.Rows, row)
	}
	return *table
}

func generateNodeData(nodes []Node) string {
	var result string
	for i, aNode := range nodes {
		result += aNode.Name
		if i < len(nodes)-1 {
			result += ","
		}
	}
	return result
}

func printAllocatedResources(ctx context.Context, k8sClient client.Client, nodeName string) (Allocated, error) {
	// List all pods on the specified node
	var podList corev1.PodList
	if err := k8sClient.List(ctx, &podList, client.MatchingFields{"spec.nodeName": nodeName}); err != nil {
		return Allocated{}, fmt.Errorf("failed to list pods on node %s.", nodeName)
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

	allocated := Allocated{
		Memory: totalMemory.String(),
		Cpu:    totalCPU.String(),
	}

	return allocated, nil
}

func findExternalHTTPSPort(cli client.Client, clusterName string) (int32, error) {
	service := corev1.Service{}
	namespacedName := types.NamespacedName{
		Name:      "ingress-nginx-controller",
		Namespace: "ingress-nginx",
	}
	err := cli.Get(context.TODO(), namespacedName, &service)
	if err != nil {
		return 0, fmt.Errorf("failed to get the ingress service on the cluster. %w", err)
	}

	localBuild := v1alpha1.Localbuild{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterName,
		},
	}
	err = cli.Get(context.TODO(), client.ObjectKeyFromObject(&localBuild), &localBuild)
	if err != nil {
		return 0, fmt.Errorf("failed to get the localbuild on the cluster. %w", err)
	}

	var targetPort corev1.ServicePort
	protocol := localBuild.Spec.BuildCustomization.Protocol + "-"
	for _, port := range service.Spec.Ports {
		if port.Name != "" && strings.HasPrefix(port.Name, protocol) {
			targetPort = port
			break
		}
	}
	return targetPort.Port, nil
}

func findInternalKubeApiPort(cli client.Client) (int32, error) {
	service := corev1.Service{}
	namespacedName := types.NamespacedName{
		Name:      "kubernetes",
		Namespace: "default",
	}
	err := cli.Get(context.TODO(), namespacedName, &service)
	if err != nil {
		return 0, fmt.Errorf("failed to get the kubernetes default service on the cluster. %w", err)
	}

	var targetPort corev1.ServicePort
	for _, port := range service.Spec.Ports {
		if port.Name != "" && strings.HasPrefix(port.Name, "https") {
			targetPort = port
			break
		}
	}
	return targetPort.TargetPort.IntVal, nil
}

// findClusterByName searches for a cluster by name in the kubeconfig
func findClusterByName(config *api.Config, name string) (*api.Cluster, bool) {
	cluster, exists := config.Clusters[name]
	return cluster, exists
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
				return nil, fmt.Errorf("Failed to build client for context %s.", contextName)
			}

			cl, err := client.New(cfg, client.Options{Scheme: k8s.GetScheme()})
			if err != nil {
				return nil, fmt.Errorf("failed to create client for context %s", contextName)
			}

			manager.clients[contextName] = cl
		}

	}
	return manager, nil
}
