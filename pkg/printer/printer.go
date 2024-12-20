package printer

import (
	"encoding/json"
	"fmt"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
	"sigs.k8s.io/yaml"
)

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
