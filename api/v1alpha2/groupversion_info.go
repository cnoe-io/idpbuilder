// +kubebuilder:object:generate=true
// +groupName=idpbuilder.cnoe.io
package v1alpha2

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "idpbuilder.cnoe.io", Version: "v1alpha2"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func init() {
	SchemeBuilder.Register(&Platform{}, &PlatformList{})
	SchemeBuilder.Register(&GiteaProvider{}, &GiteaProviderList{})
	SchemeBuilder.Register(&NginxGateway{}, &NginxGatewayList{})
	SchemeBuilder.Register(&ArgoCDProvider{}, &ArgoCDProviderList{})
}
