package version

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var (
	// Flags
	outputFormat string
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print idpbuilder version and environment info",
	Long:  "Print idpbulider version and environment info. This is useful in bug reports and CI.",
	RunE:  version,
}

func init() {
	VersionCmd.Flags().StringVarP(&outputFormat, "output", "o", "", `Print the idpbuilder version information in a given output format. Accepts "wide", "json", and "yaml".`)
}

var (
	idpbuilderVersion = "unknown"
	goVersion         = runtime.Version()
	goOs              = runtime.GOOS
	goArch            = runtime.GOARCH
	gitCommit         = "$Format:%H$"          // sha1 from git, output of $(git rev-parse HEAD)
	buildDate         = "1970-01-01T00:00:00Z" // build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
)

type idpbuilderInfo struct {
	IdpbuilderVersion string `json:"idpbuilderVersion"`
	GoVersion         string `json:"goVersion"`
	GoOs              string `json:"goOs"`
	GoArch            string `json:"goArch"`
	GitCommit         string `json:"gitCommit"`
	BuildDate         string `json:"buildDate"`
}

func version(cmd *cobra.Command, args []string) error {
	switch outputFormat {
	case "wide":
		cmd.Println(fmt.Sprintf("Version: %#v", idpbuilderInfo{
			idpbuilderVersion,
			goVersion,
			goOs,
			goArch,
			gitCommit,
			buildDate,
		}))
	case "json":
		jsonInfo, err := jsonInfo()
		if err != nil {
			return err
		}
		cmd.Println(jsonInfo)
	case "yaml":
		yamlInfo, err := yamlInfo()
		if err != nil {
			return err
		}
		cmd.Println(yamlInfo)
	case "":
		cmd.Println(fmt.Sprintf("idpbuilder %s %s %s/%s",
			idpbuilderVersion,
			goVersion,
			goOs,
			goArch))
	default:
		return fmt.Errorf("invalid output format: %s", outputFormat)
	}

	return nil
}

func jsonInfo() (string, error) {
	info := idpbuilderInfo{
		IdpbuilderVersion: idpbuilderVersion,
		GoVersion:         goVersion,
		GoOs:              goOs,
		GoArch:            goArch,
		GitCommit:         gitCommit,
		BuildDate:         buildDate,
	}
	bytes, err := json.Marshal(info)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func yamlInfo() (string, error) {
	info := idpbuilderInfo{
		IdpbuilderVersion: idpbuilderVersion,
		GoVersion:         goVersion,
		GoOs:              goOs,
		GoArch:            goArch,
		GitCommit:         gitCommit,
		BuildDate:         buildDate,
	}
	bytes, err := yaml.Marshal(info)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
