package provider

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestGetGatewayProviderStatus(t *testing.T) {
	tests := []struct {
		name    string
		obj     *unstructured.Unstructured
		want    *GatewayProviderStatus
		wantErr bool
	}{
		{
			name:    "nil object",
			obj:     nil,
			want:    nil,
			wantErr: true,
		},
		{
			name: "complete status",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"ingressClassName":     "nginx",
						"loadBalancerEndpoint": "http://172.18.0.2",
						"internalEndpoint":     "http://ingress-nginx-controller.ingress-nginx.svc.cluster.local",
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
							},
						},
					},
				},
			},
			want: &GatewayProviderStatus{
				IngressClassName:     "nginx",
				LoadBalancerEndpoint: "http://172.18.0.2",
				InternalEndpoint:     "http://ingress-nginx-controller.ingress-nginx.svc.cluster.local",
				Ready:                true,
			},
			wantErr: false,
		},
		{
			name: "partial status - no load balancer",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"ingressClassName": "nginx",
						"internalEndpoint": "http://ingress-nginx-controller.ingress-nginx.svc.cluster.local",
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "False",
							},
						},
					},
				},
			},
			want: &GatewayProviderStatus{
				IngressClassName: "nginx",
				InternalEndpoint: "http://ingress-nginx-controller.ingress-nginx.svc.cluster.local",
				Ready:            false,
			},
			wantErr: false,
		},
		{
			name: "not ready - missing Ready condition",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"ingressClassName": "nginx",
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Installing",
								"status": "True",
							},
						},
					},
				},
			},
			want: &GatewayProviderStatus{
				IngressClassName: "nginx",
				Ready:            false,
			},
			wantErr: false,
		},
		{
			name: "empty status",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{},
			},
			want: &GatewayProviderStatus{
				Ready: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetGatewayProviderStatus(tt.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGatewayProviderStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.IngressClassName != tt.want.IngressClassName {
				t.Errorf("GetGatewayProviderStatus().IngressClassName = %v, want %v", got.IngressClassName, tt.want.IngressClassName)
			}
			if got.LoadBalancerEndpoint != tt.want.LoadBalancerEndpoint {
				t.Errorf("GetGatewayProviderStatus().LoadBalancerEndpoint = %v, want %v", got.LoadBalancerEndpoint, tt.want.LoadBalancerEndpoint)
			}
			if got.InternalEndpoint != tt.want.InternalEndpoint {
				t.Errorf("GetGatewayProviderStatus().InternalEndpoint = %v, want %v", got.InternalEndpoint, tt.want.InternalEndpoint)
			}
			if got.Ready != tt.want.Ready {
				t.Errorf("GetGatewayProviderStatus().Ready = %v, want %v", got.Ready, tt.want.Ready)
			}
		})
	}
}

func TestIsGatewayProviderReady(t *testing.T) {
	tests := []struct {
		name    string
		obj     *unstructured.Unstructured
		want    bool
		wantErr bool
	}{
		{
			name:    "nil object",
			obj:     nil,
			want:    false,
			wantErr: true,
		},
		{
			name: "ready provider",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "True",
							},
						},
					},
				},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "not ready provider",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": "False",
							},
						},
					},
				},
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsGatewayProviderReady(tt.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsGatewayProviderReady() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsGatewayProviderReady() = %v, want %v", got, tt.want)
			}
		})
	}
}
