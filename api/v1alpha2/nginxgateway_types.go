package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NginxGatewaySpec defines the desired state of NginxGateway
type NginxGatewaySpec struct {
	// Namespace is the namespace where Nginx Ingress Controller will be deployed
	// +kubebuilder:validation:Required
	// +kubebuilder:default=ingress-nginx
	Namespace string `json:"namespace"`

	// Version is the version of Nginx Ingress Controller to install
	// +optional
	// +kubebuilder:default="1.13.0"
	Version string `json:"version,omitempty"`

	// IngressClass defines the ingress class configuration
	// +optional
	IngressClass NginxIngressClass `json:"ingressClass,omitempty"`
}

// NginxIngressClass defines the ingress class configuration
type NginxIngressClass struct {
	// Name is the name of the ingress class
	// +optional
	// +kubebuilder:default=nginx
	Name string `json:"name,omitempty"`

	// IsDefault indicates if this should be the default ingress class
	// +optional
	// +kubebuilder:default=true
	IsDefault bool `json:"isDefault,omitempty"`
}

// NginxGatewayStatus defines the observed state of NginxGateway
type NginxGatewayStatus struct {
	// Conditions represent the latest available observations of the NginxGateway's state
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// IngressClassName is the name of the ingress class to use in Ingress resources
	// This is a duck-typed field that all Gateway providers must expose
	// +optional
	IngressClassName string `json:"ingressClassName,omitempty"`

	// LoadBalancerEndpoint is the external endpoint for accessing services
	// This is a duck-typed field that all Gateway providers must expose
	// +optional
	LoadBalancerEndpoint string `json:"loadBalancerEndpoint,omitempty"`

	// InternalEndpoint is the cluster-internal API endpoint
	// This is a duck-typed field that all Gateway providers must expose
	// +optional
	InternalEndpoint string `json:"internalEndpoint,omitempty"`

	// Installed indicates whether Nginx has been installed
	// +optional
	Installed bool `json:"installed,omitempty"`

	// Version is the currently installed version of Nginx
	// +optional
	Version string `json:"version,omitempty"`

	// Phase represents the current phase of the Nginx gateway (e.g., Pending, Installing, Ready, Failed)
	// +optional
	Phase string `json:"phase,omitempty"`

	// Controller contains information about the Nginx controller deployment
	// +optional
	Controller NginxControllerStatus `json:"controller,omitempty"`
}

// NginxControllerStatus contains status information about the Nginx controller
type NginxControllerStatus struct {
	// Replicas is the desired number of replicas
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// ReadyReplicas is the number of ready replicas
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`
}

// NginxGateway is the Schema for the nginxgateways API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="IngressClass",type=string,JSONPath=`.status.ingressClassName`
// +kubebuilder:printcolumn:name="LoadBalancer",type=string,JSONPath=`.status.loadBalancerEndpoint`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type NginxGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NginxGatewaySpec   `json:"spec,omitempty"`
	Status NginxGatewayStatus `json:"status,omitempty"`
}

// NginxGatewayList contains a list of NginxGateway
// +kubebuilder:object:root=true
type NginxGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NginxGateway `json:"items"`
}
