package printer

import (
	"encoding/json"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
	"sigs.k8s.io/yaml"
)

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
