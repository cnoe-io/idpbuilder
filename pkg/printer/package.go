package printer

import (
	"fmt"
	"github.com/cnoe-io/idpbuilder/pkg/entity"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PackagePrinter struct {
	Packages  []entity.Package
	OutWriter io.Writer
}

func (pp PackagePrinter) PrintOutput(format string) error {
	switch format {
	case "json":
		return PrintDataAsJson(pp.Packages, pp.OutWriter)
	case "yaml":
		return PrintDataAsYaml(pp.Packages, pp.OutWriter)
	case "table":
		return PrintDataAsTable(generatePackageTable(pp.Packages), pp.OutWriter)
	default:
		return fmt.Errorf("output format %s is not supported", format)
	}
}

func generatePackageTable(packagesTable []entity.Package) metav1.Table {
	table := &metav1.Table{}
	table.ColumnDefinitions = []metav1.TableColumnDefinition{
		{Name: "Custom package name", Type: "string"},
		{Name: "idp namespace", Type: "string"},
		{Name: "Git Repository", Type: "string"},
		{Name: "Status", Type: "string"},
	}
	for _, p := range packagesTable {
		row := metav1.TableRow{
			Cells: []interface{}{
				p.Name,
				p.Namespace,
				p.GitRepository,
				p.Status,
			},
		}
		table.Rows = append(table.Rows, row)
	}
	return *table
}
