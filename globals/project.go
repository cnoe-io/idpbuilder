package globals

import "fmt"

const ProjectName string = "idpbuilder"

func GetProjectNamespace(name string) string {
	return fmt.Sprintf("%s-%s", ProjectName, name)
}
