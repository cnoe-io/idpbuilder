package util

type CorePackageTemplateConfig struct {
	Protocol       string
	Host           string
	IngressHost    string
	Port           string
	UsePathRouting bool
	SelfSignedCert string
}
