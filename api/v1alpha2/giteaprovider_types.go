package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GiteaProviderSpec defines the desired state of GiteaProvider
type GiteaProviderSpec struct {
	// Namespace is the namespace where Gitea will be deployed
	// +kubebuilder:validation:Required
	// +kubebuilder:default=gitea
	Namespace string `json:"namespace"`

	// Version is the version of Gitea to install
	// +optional
	// +kubebuilder:default="1.24.3"
	Version string `json:"version,omitempty"`

	// Protocol is the protocol to use for Gitea endpoint (http or https)
	// +optional
	// +kubebuilder:default="http"
	Protocol string `json:"protocol,omitempty"`

	// Host is the hostname for Gitea endpoint
	// +optional
	// +kubebuilder:default="cnoe.localtest.me"
	Host string `json:"host,omitempty"`

	// Port is the port for Gitea endpoint
	// +optional
	// +kubebuilder:default="8080"
	Port string `json:"port,omitempty"`

	// UsePathRouting indicates whether to use path-based routing
	// +optional
	// +kubebuilder:default=false
	UsePathRouting bool `json:"usePathRouting,omitempty"`

	// AdminUser defines the Gitea admin user configuration
	// +optional
	AdminUser GiteaAdminUser `json:"adminUser,omitempty"`

	// Organizations is a list of Gitea organizations to create
	// +optional
	Organizations []GiteaOrganization `json:"organizations,omitempty"`
}

// GiteaAdminUser defines the admin user configuration for Gitea
type GiteaAdminUser struct {
	// Username is the admin username
	// +optional
	// +kubebuilder:default="giteaAdmin"
	Username string `json:"username,omitempty"`

	// Email is the admin user email
	// +optional
	// +kubebuilder:default="admin@cnoe.localtest.me"
	Email string `json:"email,omitempty"`

	// PasswordSecretRef references a secret containing the admin password
	// +optional
	PasswordSecretRef *SecretReference `json:"passwordSecretRef,omitempty"`

	// AutoGenerate indicates whether to auto-generate credentials if not provided
	// +optional
	// +kubebuilder:default=true
	AutoGenerate bool `json:"autoGenerate,omitempty"`
}

// GiteaOrganization defines a Gitea organization to create
type GiteaOrganization struct {
	// Name is the organization name
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Description is the organization description
	// +optional
	Description string `json:"description,omitempty"`
}

// SecretReference references a Kubernetes Secret
type SecretReference struct {
	// Name is the name of the secret
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace is the namespace of the secret
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`

	// Key is the key within the secret
	// +kubebuilder:validation:Required
	Key string `json:"key"`
}

// GiteaProviderStatus defines the observed state of GiteaProvider
type GiteaProviderStatus struct {
	// Conditions represent the latest available observations of the GiteaProvider's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Endpoint is the external URL for Gitea web UI and cloning
	// This is a duck-typed field that all Git providers must expose
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// InternalEndpoint is the cluster-internal URL for Gitea API access
	// This is a duck-typed field that all Git providers must expose
	// +optional
	InternalEndpoint string `json:"internalEndpoint,omitempty"`

	// CredentialsSecretRef references the secret containing Gitea credentials
	// This is a duck-typed field that all Git providers must expose
	// +optional
	CredentialsSecretRef *SecretReference `json:"credentialsSecretRef,omitempty"`

	// Installed indicates whether Gitea has been installed
	// +optional
	Installed bool `json:"installed,omitempty"`

	// Version is the currently installed version of Gitea
	// +optional
	Version string `json:"version,omitempty"`

	// Phase represents the current phase of the Gitea provider (e.g., Pending, Installing, Ready, Failed)
	// +optional
	Phase string `json:"phase,omitempty"`

	// AdminUser contains information about the admin user
	// +optional
	AdminUser GiteaAdminUserStatus `json:"adminUser,omitempty"`
}

// GiteaAdminUserStatus contains status information about the Gitea admin user
type GiteaAdminUserStatus struct {
	// Username is the admin username
	// +optional
	Username string `json:"username,omitempty"`

	// SecretRef references the secret containing admin credentials
	// +optional
	SecretRef *SecretReference `json:"secretRef,omitempty"`
}

// GiteaProvider is the Schema for the giteaproviders API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=`.status.endpoint`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type GiteaProvider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GiteaProviderSpec   `json:"spec,omitempty"`
	Status GiteaProviderStatus `json:"status,omitempty"`
}

// GiteaProviderList contains a list of GiteaProvider
// +kubebuilder:object:root=true
type GiteaProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GiteaProvider `json:"items"`
}
