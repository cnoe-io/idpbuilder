package globals

import "fmt"

const (
	ProjectName string = "idpbuilder"

	ArgoCDNamespace  string = "argocd"
	TraefikNamespace string = "traefik"

	SelfSignedCertSecretName = "idpbuilder-cert"
	SelfSignedCertCMName     = "idpbuilder-cert"
	SelfSignedCertCMKeyName  = "ca.crt"
	DefaultSANWildcard       = "*.cnoe.localtest.me"
	DefaultHostName          = "cnoe.localtest.me"
)

func GetProjectNamespace(name string) string {
	return fmt.Sprintf("%s-%s", ProjectName, name)
}
