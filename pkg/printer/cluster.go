package printer

import (
	"fmt"
	"github.com/cnoe-io/idpbuilder/pkg/printer/types"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterPrinter struct {
	Clusters  []types.Cluster
	OutWriter io.Writer
}

func (cp ClusterPrinter) PrintOutput(format string) error {
	switch format {
	case "json":
		return PrintDataAsJson(cp.Clusters, cp.OutWriter)
	case "yaml":
		return PrintDataAsYaml(cp.Clusters, cp.OutWriter)
	case "table":
		return PrintDataAsTable(generateClusterTable(cp.Clusters), cp.OutWriter)
	default:
		return fmt.Errorf("output format %s is not supported", format)
	}
}

func generateClusterTable(input []types.Cluster) metav1.Table {
	table := &metav1.Table{}
	table.ColumnDefinitions = []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string"},
		{Name: "External-Port", Type: "string"},
		{Name: "Kube-Api", Type: "string"},
		{Name: "TLS", Type: "string"},
		{Name: "Kube-Port", Type: "string"},
		{Name: "Nodes", Type: "string"},
	}

	for _, cluster := range input {
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

func generateNodeData(nodes []types.Node) string {
	var result string
	for i, aNode := range nodes {
		result += aNode.Name
		if i < len(nodes)-1 {
			result += ","
		}
	}
	return result
}
