package globals

import "fmt"

const (
	ProjectName string = "idpbuilder"

	NginxNamespace  string = "ingress-nginx"
	ArgoCDNamespace string = "argocd"

	SelfSignedCertSecretName = "idpbuilder-cert"
	SelfSignedCertCMName     = "idpbuilder-cert"
	SelfSignedCertCMKeyName  = "ca.crt"
	DefaultSANWildcard       = "*.cnoe.localtest.me"
	DefaultHostName          = "cnoe.localtest.me"
)

var (
	ClusterProtocol string // http or https scheme
	ClusterPort     string // 8443 or user's port defined when the cluster is created
)

func GetProjectNamespace(name string) string {
	return fmt.Sprintf("%s-%s", ProjectName, name)
}
