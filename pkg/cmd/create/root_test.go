package create

import (
	"testing"

	"github.com/cnoe-io/idpbuilder/api/v1alpha1"
)

func TestParseRegistryMirrors(t *testing.T) {
	type test struct {
		name    string
		input   []string
		expect  []v1alpha1.RegistryMirror
		wantErr bool
	}

	tests := []test{
		{
			name:    "empty input",
			input:   []string{},
			expect:  []v1alpha1.RegistryMirror{},
			wantErr: false,
		},
		{
			name:  "single mirror",
			input: []string{"docker.io=http://kind-registry:5000"},
			expect: []v1alpha1.RegistryMirror{
				{
					TargetRegistry:  "docker.io",
					RegistryAddress: "http://kind-registry:5000",
				},
			},
			wantErr: false,
		},
		{
			name:  "multiple mirrors",
			input: []string{"docker.io=http://kind-registry:5000", "ghcr.io=http://kind-registry:5000"},
			expect: []v1alpha1.RegistryMirror{
				{
					TargetRegistry:  "docker.io",
					RegistryAddress: "http://kind-registry:5000",
				},
				{
					TargetRegistry:  "ghcr.io",
					RegistryAddress: "http://kind-registry:5000",
				},
			},
			wantErr: false,
		},
		{
			name:  "mirror with space",
			input: []string{" docker.io = http://kind-registry:5000 "},
			expect: []v1alpha1.RegistryMirror{
				{
					TargetRegistry:  "docker.io",
					RegistryAddress: "http://kind-registry:5000",
				},
			},
			wantErr: false,
		},
		{
			name:    "missing equals",
			input:   []string{"docker.io:http://kind-registry:5000"},
			expect:  nil,
			wantErr: true,
		},
		{
			name:    "empty target",
			input:   []string{"=http://kind-registry:5000"},
			expect:  nil,
			wantErr: true,
		},
		{
			name:    "empty address",
			input:   []string{"docker.io="},
			expect:  nil,
			wantErr: true,
		},
		{
			name:    "target with path",
			input:   []string{"docker.io/library=http://my-registry:5000"},
			wantErr: true,
		},
		{
			name:    "target with scheme",
			input:   []string{"https://docker.io=http://my-registry:5000"},
			wantErr: true,
		},
		{
			name:    "malformed address URL",
			input:   []string{"docker.io=not-a-url"},
			wantErr: true,
		},
		{
			name:  "mirror with https",
			input: []string{"docker.io=https://my-registry:5000"},
			expect: []v1alpha1.RegistryMirror{
				{
					TargetRegistry:  "docker.io",
					RegistryAddress: "https://my-registry:5000",
				},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := parseRegistryMirrors(tc.input)
			if (err != nil) != tc.wantErr {
				t.Errorf("parseRegistryMirrors() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if !tc.wantErr {
				if len(result) != len(tc.expect) {
					t.Errorf("parseRegistryMirrors() got %d mirrors, want %d", len(result), len(tc.expect))
					return
				}
				for i := range result {
					if result[i].TargetRegistry != tc.expect[i].TargetRegistry {
						t.Errorf("parseRegistryMirrors()[%d].TargetRegistry = %v, want %v", i, result[i].TargetRegistry, tc.expect[i].TargetRegistry)
					}
					if result[i].RegistryAddress != tc.expect[i].RegistryAddress {
						t.Errorf("parseRegistryMirrors()[%d].RegistryAddress = %v, want %v", i, result[i].RegistryAddress, tc.expect[i].RegistryAddress)
					}
				}
			}
		})
	}
}
