package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PlatformSpec defines the desired state of Platform
type PlatformSpec struct {
	// Domain is the base domain for the platform
	// +kubebuilder:validation:Required
	Domain string `json:"domain"`

	// Components defines the platform component configuration
	// +kubebuilder:validation:Required
	Components PlatformComponents `json:"components"`
}

// PlatformComponents defines the components that make up the platform
type PlatformComponents struct {
	// GitProviders is a list of Git provider references
	// +optional
	GitProviders []ProviderReference `json:"gitProviders,omitempty"`

	// Gateways is a list of Gateway provider references
	// +optional
	Gateways []ProviderReference `json:"gateways,omitempty"`

	// GitOpsProviders is a list of GitOps provider references
	// +optional
	GitOpsProviders []ProviderReference `json:"gitOpsProviders,omitempty"`
}

// ProviderReference references a provider CR by name, kind, and namespace
type ProviderReference struct {
	// Name is the name of the provider CR
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Kind is the kind of the provider CR (e.g., GiteaProvider, NginxGateway)
	// +kubebuilder:validation:Required
	Kind string `json:"kind"`

	// Namespace is the namespace of the provider CR
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`
}

// PlatformStatus defines the observed state of Platform
type PlatformStatus struct {
	// Conditions represent the latest available observations of the Platform's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Providers contains the aggregated status of all provider references
	// +optional
	Providers PlatformProviderStatus `json:"providers,omitempty"`

	// ObservedGeneration reflects the generation of the most recently observed Platform
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Phase represents the current phase of the Platform (e.g., Pending, Ready, Failed)
	// +optional
	Phase string `json:"phase,omitempty"`
}

// PlatformProviderStatus contains aggregated status from all providers
type PlatformProviderStatus struct {
	// GitProviders contains status of Git providers
	// +optional
	GitProviders []ProviderStatusSummary `json:"gitProviders,omitempty"`

	// Gateways contains status of Gateway providers
	// +optional
	Gateways []ProviderStatusSummary `json:"gateways,omitempty"`

	// GitOpsProviders contains status of GitOps providers
	// +optional
	GitOpsProviders []ProviderStatusSummary `json:"gitOpsProviders,omitempty"`
}

// ProviderStatusSummary summarizes the status of a provider
type ProviderStatusSummary struct {
	// Name is the name of the provider
	Name string `json:"name"`

	// Kind is the kind of the provider
	Kind string `json:"kind"`

	// Ready indicates whether the provider is ready
	Ready bool `json:"ready"`
}

// Platform is the Schema for the platforms API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Platform struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PlatformSpec   `json:"spec,omitempty"`
	Status PlatformStatus `json:"status,omitempty"`
}

// PlatformList contains a list of Platform
// +kubebuilder:object:root=true
type PlatformList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Platform `json:"items"`
}
