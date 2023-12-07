package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	//// Type specifies what kind of package this is. local means local files specified in the Directory field must be synced to a Git Repository
	//Type string `json:"type"`
	// Replicate specifies whether to replicate remote or local contents to the local gitea server.
	Replicate bool `json:"replicate"`
	// GitServerURL specifies the base URL for the git server for API calls. for example, http://gitea.cnoe.localtest.me:8880
	GitServerURL           string          `json:"gitServerURL"`
	GitServerAuthSecretRef SecretReference `json:"gitServerAuthSecretRef"`

	ArgoCD ArgoCDPackageSpec `json:"argoCD,omitempty"`
}

type ArgoCDPackageSpec struct {
	// ApplicationFile specifies the absolute path to the ArgoCD application file
	ApplicationFile string `json:"applicationFile"`
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
}

type CustomPackageStatus struct {
	Synced bool `json:"synced,omitempty"`
}
