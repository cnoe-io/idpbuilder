package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GitServerSource struct {
	// Emedded enables the image containing the go binary embedded contents
	Embedded bool `json:"embedded,omitempty"`

	// Image specifies a docker image to use. Specifying this disables installation of embedded applications.
	Image string `json:"image,omitempty"`
}

type GitServerSpec struct {
	Source GitServerSource `json:"source,omitempty"`
}

type GitServerStatus struct {
	// ObservedGeneration is the 'Generation' of the Service that was last processed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// ImageID is the id of the most recently generated docker image or empty if no image has been created
	ImageID string `json:"imageID,omitempty"`

	// Host is the host value of the ingress rule
	Host string `json:"host,omitempty"`

	DeploymentAvailable bool `json:"deploymentAvailable"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type GitServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitServerSpec   `json:"spec,omitempty"`
	Status GitServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type GitServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitServer `json:"items"`
}
