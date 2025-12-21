package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ArgoCDProviderSpec defines the desired state of ArgoCDProvider
type ArgoCDProviderSpec struct {
	// Namespace is the namespace where ArgoCD will be deployed
	// +kubebuilder:validation:Required
	// +kubebuilder:default=argocd
	Namespace string `json:"namespace"`

	// Version is the version of ArgoCD to install
	// +optional
	// +kubebuilder:default="v2.12.0"
	Version string `json:"version,omitempty"`

	// AdminCredentials defines the ArgoCD admin credentials configuration
	// +optional
	AdminCredentials ArgoCDAdminCredentials `json:"adminCredentials,omitempty"`

	// Projects is a list of ArgoCD projects to create
	// +optional
	Projects []ArgoCDProject `json:"projects,omitempty"`
}

// ArgoCDAdminCredentials defines the admin credentials configuration for ArgoCD
type ArgoCDAdminCredentials struct {
	// SecretRef references a secret containing the admin credentials
	// +optional
	SecretRef *SecretReference `json:"secretRef,omitempty"`

	// AutoGenerate indicates whether to auto-generate credentials if not provided
	// +optional
	// +kubebuilder:default=true
	AutoGenerate bool `json:"autoGenerate,omitempty"`
}

// ArgoCDProject defines an ArgoCD project to create
type ArgoCDProject struct {
	// Name is the project name
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Description is the project description
	// +optional
	Description string `json:"description,omitempty"`
}

// ArgoCDProviderStatus defines the observed state of ArgoCDProvider
type ArgoCDProviderStatus struct {
	// Conditions represent the latest available observations of the ArgoCDProvider's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Endpoint is the external URL for ArgoCD web UI
	// This is a duck-typed field that all GitOps providers must expose
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// InternalEndpoint is the cluster-internal URL for ArgoCD API access
	// This is a duck-typed field that all GitOps providers must expose
	// +optional
	InternalEndpoint string `json:"internalEndpoint,omitempty"`

	// CredentialsSecretRef references the secret containing ArgoCD admin credentials
	// This is a duck-typed field that all GitOps providers must expose
	// +optional
	CredentialsSecretRef *SecretReference `json:"credentialsSecretRef,omitempty"`

	// Installed indicates whether ArgoCD has been installed
	// +optional
	Installed bool `json:"installed,omitempty"`

	// Version is the currently installed version of ArgoCD
	// +optional
	Version string `json:"version,omitempty"`

	// Phase represents the current phase of the ArgoCD provider (e.g., Pending, Installing, Ready, Failed)
	// +optional
	Phase string `json:"phase,omitempty"`
}

// ArgoCDProvider is the Schema for the argocdproviders API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=`.status.endpoint`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type ArgoCDProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArgoCDProviderSpec   `json:"spec,omitempty"`
	Status ArgoCDProviderStatus `json:"status,omitempty"`
}

// ArgoCDProviderList contains a list of ArgoCDProvider
// +kubebuilder:object:root=true
type ArgoCDProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArgoCDProvider `json:"items"`
}
