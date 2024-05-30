package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	CNOEURIScheme = "cnoe://"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type CustomPackage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomPackageSpec   `json:"spec,omitempty"`
	Status CustomPackageStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type CustomPackageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomPackage `json:"items"`
}

// CustomPackageSpec controls the installation of the custom applications.
type CustomPackageSpec struct {
	ArgoCD ArgoCDPackageSpec `json:"argoCD,omitempty"`
	// GitServerURL specifies the base URL for the git server for API calls.
	// for example, https://gitea.cnoe.localtest.me:8443
	GitServerURL           string          `json:"gitServerURL"`
	GitServerAuthSecretRef SecretReference `json:"gitServerAuthSecretRef"`
	// InternalGitServeURL specifies the base URL for the git server accessible within the cluster.
	// for example, http://my-gitea-http.gitea.svc.cluster.local:3000
	InternalGitServeURL string               `json:"internalGitServeURL"`
	RemoteRepository    RemoteRepositorySpec `json:"remoteRepository"`
	// Replicate specifies whether to replicate remote or local contents to the local gitea server.
	// +kubebuilder:default:=false
	Replicate bool `json:"replicate"`
}

// RemoteRepositorySpec specifies information about remote repositories.
type RemoteRepositorySpec struct {
	CloneSubmodules bool   `json:"cloneSubmodules"`
	Path            string `json:"path"`
	// Url specifies the url to the repository containing the ArgoCD application file
	Url string `json:"url"`
	// Ref specifies the specific ref supported by git fetch
	Ref string `json:"ref"`
}

type ArgoCDPackageSpec struct {
	// ApplicationFile specifies the absolute path to the ArgoCD application file
	ApplicationFile string `json:"applicationFile"`
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
}

type CustomPackageStatus struct {
	// A Custom package is considered synced when the in-cluster repository url is set as the repository URL
	// This only applies for a package that references local directories
	Synced            bool        `json:"synced,omitempty"`
	GitRepositoryRefs []ObjectRef `json:"gitRepositoryRefs,omitempty"`
}

type ObjectRef struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Name       string `json:"name,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	Kind       string `json:"kind,omitempty"`
	UID        string `json:"uid,omitempty"`
}
