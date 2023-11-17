package globals

import "fmt"

const ProjectName string = "idpbuilder"
const giteaResourceName string = "gitea"
const gitServerResourceName string = "gitserver"

func GetProjectNamespace(name string) string {
	return fmt.Sprintf("%s-%s", ProjectName, name)
}

func GiteaResourceName() string {
	return giteaResourceName
}

func GitServerResourcename() string {
	return gitServerResourceName
}
