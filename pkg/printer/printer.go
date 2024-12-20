package printer

import (
	"encoding/json"
	"fmt"
	"github.com/cnoe-io/idpbuilder/pkg/entity"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
	"sigs.k8s.io/yaml"
	"strings"
)

type Printer interface {
	PrintOutput()
}

type ClusterPrinter struct {
	Clusters  []entity.Cluster
	OutWriter io.Writer
}

type SecretPrinter struct {
	Secrets   []entity.Secret
	OutWriter io.Writer
}

func (sp SecretPrinter) PrintOutput(format string) error {
	switch format {
	case "json":
		return PrintDataAsJson(sp.Secrets, sp.OutWriter)
	case "yaml":
		return PrintDataAsYaml(sp.Secrets, sp.OutWriter)
	case "table":
		return PrintDataAsTable(sp.generateSecretTable(sp.Secrets), sp.OutWriter)
	default:
		return fmt.Errorf("output format %s is not supported", format)
	}
}

func (sp SecretPrinter) generateSecretTable(secretTable []entity.Secret) metav1.Table {
	table := &metav1.Table{}
	table.ColumnDefinitions = []metav1.TableColumnDefinition{
		{Name: "Name", Type: "string"},
		{Name: "Namespace", Type: "string"},
		{Name: "Username", Type: "string"},
		{Name: "Password", Type: "string"},
		{Name: "Token", Type: "string"},
		{Name: "Data", Type: "string"},
	}
	for _, secret := range secretTable {
		var dataEntries []string

		if !secret.IsCore {
			for key, value := range secret.Data {
				dataEntries = append(dataEntries, fmt.Sprintf("%s=%s", key, value))
			}
		}
		dataString := strings.Join(dataEntries, ", ")
		row := metav1.TableRow{
			Cells: []interface{}{
				secret.Name,
				secret.Namespace,
				secret.Username,
				secret.Password,
				secret.Token,
				dataString,
			},
		}
		table.Rows = append(table.Rows, row)
	}
	return *table
}

func (cp ClusterPrinter) PrintOutput(format string) error {
	switch format {
	case "json":
		return PrintDataAsJson(cp.Clusters, cp.OutWriter)
	case "yaml":
		return PrintDataAsYaml(cp.Clusters, cp.OutWriter)
	case "table":
		return PrintDataAsTable(cp.generateClusterTable(cp.Clusters), cp.OutWriter)
	default:
		return fmt.Errorf("output format %s is not supported", format)
	}
}

func (cp ClusterPrinter) generateClusterTable(input []entity.Cluster) metav1.Table {
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

func generateNodeData(nodes []entity.Node) string {
	var result string
	for i, aNode := range nodes {
		result += aNode.Name
		if i < len(nodes)-1 {
			result += ","
		}
	}
	return result
}

func PrintOutput[T any](outWriter io.Writer, input []T, inputTable metav1.Table, format string) error {
	switch format {
	case "json":
		return PrintDataAsJson(input, outWriter)
	case "yaml":
		return PrintDataAsYaml(input, outWriter)
	case "table":
		return PrintDataAsTable(inputTable, outWriter)
	default:
		return fmt.Errorf("output format %s is not supported", format)
	}
}

func PrintDataAsTable(table metav1.Table, outWriter io.Writer) error {
	printer := printers.NewTablePrinter(printers.PrintOptions{})
	return printer.PrintObj(&table, outWriter)
}

func PrintDataAsJson(data any, outWriter io.Writer) error {
	enc := json.NewEncoder(outWriter)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func PrintDataAsYaml(data any, outWriter io.Writer) error {
	b, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	_, err = outWriter.Write(b)
	return err
}
