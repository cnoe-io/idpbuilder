package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	GitProviderGitea   = "gitea"
	GitProviderGitHub  = "github"
	GiteaAdminUserName = "giteaAdmin"
	SourceTypeLocal    = "local"
	SourceTypeRemote   = "remote"
	SourceTypeEmbedded = "embedded"
)

type GitRepositorySpec struct {
	// +kubebuilder:validation:Optional
	Customization PackageCustomization `json:"customization,omitempty"`
	// SecretRef is the reference to secret that contain Git server credentials
	// +kubebuilder:validation:Optional
	SecretRef SecretReference     `json:"secretRef"`
	Source    GitRepositorySource `json:"source,omitempty"`
	Provider  Provider            `json:"provider"`
}

type GitRepositorySource struct {
	// +kubebuilder:validation:Enum:=argocd;gitea;nginx
	// +kubebuilder:validation:Optional
	EmbeddedAppName string `json:"embeddedAppName,omitempty"`
	// Path is the absolute path to directory that contains Kustomize structure or raw manifests.
	// This is required when Type is set to local.
	// +kubebuilder:validation:Optional
	Path             string               `json:"path"`
	RemoteRepository RemoteRepositorySpec `json:"remoteRepository"`
	// Type is the source type.
	// +kubebuilder:validation:Enum:=local;embedded;remote
	// +kubebuilder:default:=embedded
	Type string `json:"type"`
}

type Provider struct {
	// +kubebuilder:validation:Enum:=gitea;github
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// GitURL is the base URL of Git server used for API calls.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^https?:\/\/.+$`
	GitURL string `json:"gitURL"`
	// InternalGitURL is the base URL of Git server accessible within the cluster only.
	InternalGitURL   string `json:"internalGitURL"`
	OrganizationName string `json:"organizationName"`
}

type SecretReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type Commit struct {
	// Hash is the digest of the most recent commit
	// +kubebuilder:validation:Optional
	Hash string `json:"hash"`
}

type GitRepositoryStatus struct {
	// LatestCommit is the most recent commit known to the controller
	// +kubebuilder:validation:Optional
	LatestCommit Commit `json:"commit"`
	// ExternalGitRepositoryUrl is the url for the in-cluster repository accessible from local machine.
	// +kubebuilder:validation:Optional
	ExternalGitRepositoryUrl string `json:"externalGitRepositoryUrl"`
	// InternalGitRepositoryUrl is the url for the in-cluster repository accessible within the cluster.
	// +kubebuilder:validation:Optional
	InternalGitRepositoryUrl string `json:"internalGitRepositoryUrl"`
	// Path is the path within the repository that contains the files.
	// +kubebuilder:validation:Optional
	Path   string `json:"path"`
	Synced bool   `json:"synced"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type GitRepository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitRepositorySpec   `json:"spec,omitempty"`
	Status GitRepositoryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type GitRepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitRepository `json:"items"`
}
