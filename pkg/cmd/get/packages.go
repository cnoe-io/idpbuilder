package get

import (
	"github.com/spf13/cobra"
)

const (
	packageTemplatePath = "templates/package.tmpl"
)

var PackagesCmd = &cobra.Command{
	Use:   "packages",
	Short: "retrieve package info from the cluster",
	Long:  ``,
	RunE:  getPackagesE,
}

type PackageTemplateData struct {
	Name        string                  `json:"name"`
	Namespace   string                  `json:"namespace"`
	Category    string                  `json:"category"`
	PackageType string                  `json:"packageType"`
	Spec        PackageSpecTemplateData `json:"spec"`
}

type PackageSpecTemplateData struct {
	Description string            `json:"description"`
	URL         string            `json:"url"`
	Credentials map[string]string `json:"credentials"`
}

func getPackagesE(cmd *cobra.Command, args []string) error {
	return nil
}
