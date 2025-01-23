package printer

import (
	"fmt"
	"github.com/cnoe-io/idpbuilder/pkg/printer/types"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

type SecretPrinter struct {
	Secrets   []types.Secret
	OutWriter io.Writer
}

func (sp SecretPrinter) PrintOutput(format string) error {
	switch format {
	case "json":
		return PrintDataAsJson(sp.Secrets, sp.OutWriter)
	case "yaml":
		return PrintDataAsYaml(sp.Secrets, sp.OutWriter)
	case "table":
		return PrintDataAsTable(generateSecretTable(sp.Secrets), sp.OutWriter)
	default:
		return fmt.Errorf("output format %s is not supported", format)
	}
}

func generateSecretTable(secretTable []types.Secret) metav1.Table {
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
