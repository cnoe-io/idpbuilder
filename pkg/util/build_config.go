package util

type PackageTemplateConfig struct {
	Protocol       string
	Host           string
	IngressHost    string
	Port           string
	UsePathRouting bool
	SelfSignedCert string
	// Data field contains custom data and all above field name and values. Used for templating.
	// This is done to avoid end users having to nest every single custom properties when writing templates.
	Data map[string]any
}

func NewPackageTemplateConfig(protocol, host, ingressHost, port string, usePathRouting bool, customData map[string]any) PackageTemplateConfig {
	if customData == nil {
		customData = make(map[string]any, 6)
	}

	p := PackageTemplateConfig{
		Protocol:       protocol,
		Host:           host,
		IngressHost:    ingressHost,
		Port:           port,
		UsePathRouting: usePathRouting,
		Data:           customData,
	}

	p.Data["Protocol"] = p.Protocol
	p.Data["Host"] = p.Host
	p.Data["IngressHost"] = p.IngressHost
	p.Data["Port"] = p.Port
	p.Data["UsePathRouting"] = p.UsePathRouting

	return p
}
