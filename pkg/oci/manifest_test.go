package oci

import (
	"testing"
)

func TestParseManifest(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		wantType    string
		wantError   bool
		errorContains string
	}{
		{
			name: "valid OCI manifest",
			data: `{
				"schemaVersion": 2,
				"mediaType": "application/vnd.oci.image.manifest.v1+json",
				"config": {
					"mediaType": "application/vnd.oci.image.config.v1+json",
					"size": 512,
					"digest": "sha256:config123"
				},
				"layers": [
					{
						"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
						"size": 1024,
						"digest": "sha256:layer1"
					}
				]
			}`,
			wantType: "*oci.Manifest",
		},
		{
			name: "valid Docker manifest",
			data: `{
				"schemaVersion": 2,
				"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
				"config": {
					"mediaType": "application/vnd.docker.container.image.v1+json",
					"size": 512,
					"digest": "sha256:config123"
				},
				"layers": [
					{
						"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
						"size": 1024,
						"digest": "sha256:layer1"
					}
				]
			}`,
			wantType: "*oci.Manifest",
		},
		{
			name: "valid OCI manifest list",
			data: `{
				"schemaVersion": 2,
				"mediaType": "application/vnd.oci.image.index.v1+json",
				"manifests": [
					{
						"mediaType": "application/vnd.oci.image.manifest.v1+json",
						"size": 1024,
						"digest": "sha256:manifest1",
						"platform": {
							"architecture": "amd64",
							"os": "linux"
						}
					}
				]
			}`,
			wantType: "*oci.ManifestList",
		},
		{
			name:          "empty data",
			data:          "",
			wantError:     true,
			errorContains: "manifest data is empty",
		},
		{
			name:          "invalid JSON",
			data:          "{invalid json",
			wantError:     true,
			errorContains: "failed to unmarshal manifest",
		},
		{
			name: "missing mediaType",
			data: `{
				"schemaVersion": 2,
				"config": {}
			}`,
			wantError:     true,
			errorContains: "manifest missing mediaType field",
		},
		{
			name: "unsupported mediaType",
			data: `{
				"schemaVersion": 2,
				"mediaType": "application/vnd.unknown.type+json"
			}`,
			wantError:     true,
			errorContains: "unsupported media type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseManifest([]byte(tt.data))

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Errorf("expected result but got nil")
				return
			}

			// Check type
			switch tt.wantType {
			case "*oci.Manifest":
				if _, ok := result.(*Manifest); !ok {
					t.Errorf("expected *Manifest, got %T", result)
				}
			case "*oci.ManifestList":
				if _, ok := result.(*ManifestList); !ok {
					t.Errorf("expected *ManifestList, got %T", result)
				}
			}
		})
	}
}

func TestValidateManifest(t *testing.T) {
	tests := []struct {
		name        string
		manifest    *Manifest
		wantError   bool
		errorContains string
	}{
		{
			name: "valid manifest",
			manifest: &Manifest{
				SchemaVersion: SchemaVersion,
				MediaType:     MediaTypeManifest,
				Config: Descriptor{
					MediaType: MediaTypeConfig,
					Digest:    "sha256:config123",
					Size:      512,
				},
				Layers: []Descriptor{
					{
						MediaType: MediaTypeLayer,
						Digest:    "sha256:layer1",
						Size:      1024,
					},
				},
			},
			wantError: false,
		},
		{
			name:          "nil manifest",
			manifest:      nil,
			wantError:     true,
			errorContains: "manifest is nil",
		},
		{
			name: "invalid schema version",
			manifest: &Manifest{
				SchemaVersion: 1,
				MediaType:     MediaTypeManifest,
			},
			wantError:     true,
			errorContains: "unsupported schema version",
		},
		{
			name: "missing mediaType",
			manifest: &Manifest{
				SchemaVersion: SchemaVersion,
				MediaType:     "",
			},
			wantError:     true,
			errorContains: "manifest mediaType is required",
		},
		{
			name: "invalid mediaType",
			manifest: &Manifest{
				SchemaVersion: SchemaVersion,
				MediaType:     "invalid/type",
			},
			wantError:     true,
			errorContains: "invalid manifest mediaType",
		},
		{
			name: "invalid config descriptor",
			manifest: &Manifest{
				SchemaVersion: SchemaVersion,
				MediaType:     MediaTypeManifest,
				Config: Descriptor{
					// Missing required fields
				},
			},
			wantError:     true,
			errorContains: "invalid config descriptor",
		},
		{
			name: "no layers",
			manifest: &Manifest{
				SchemaVersion: SchemaVersion,
				MediaType:     MediaTypeManifest,
				Config: Descriptor{
					MediaType: MediaTypeConfig,
					Digest:    "sha256:config123",
					Size:      512,
				},
				Layers: []Descriptor{},
			},
			wantError:     true,
			errorContains: "manifest must have at least one layer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateManifest(tt.manifest)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateManifestList(t *testing.T) {
	tests := []struct {
		name          string
		manifestList  *ManifestList
		wantError     bool
		errorContains string
	}{
		{
			name: "valid manifest list",
			manifestList: &ManifestList{
				SchemaVersion: SchemaVersion,
				MediaType:     MediaTypeManifestList,
				Manifests: []Descriptor{
					{
						MediaType: MediaTypeManifest,
						Digest:    "sha256:manifest1",
						Size:      1024,
						Platform: &Platform{
							Architecture: ArchitectureAmd64,
							OS:           OSLinux,
						},
					},
				},
			},
			wantError: false,
		},
		{
			name:          "nil manifest list",
			manifestList:  nil,
			wantError:     true,
			errorContains: "manifest list is nil",
		},
		{
			name: "empty manifests",
			manifestList: &ManifestList{
				SchemaVersion: SchemaVersion,
				MediaType:     MediaTypeManifestList,
				Manifests:     []Descriptor{},
			},
			wantError:     true,
			errorContains: "manifest list must have at least one manifest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateManifestList(tt.manifestList)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCreateManifest(t *testing.T) {
	config := Descriptor{
		MediaType: MediaTypeConfig,
		Digest:    "sha256:config123",
		Size:      512,
	}

	layers := []Descriptor{
		{
			MediaType: MediaTypeLayer,
			Digest:    "sha256:layer1",
			Size:      1024,
		},
	}

	manifest := CreateManifest(config, layers)

	if manifest.SchemaVersion != SchemaVersion {
		t.Errorf("expected schema version %d, got %d", SchemaVersion, manifest.SchemaVersion)
	}

	if manifest.MediaType != MediaTypeManifest {
		t.Errorf("expected media type %s, got %s", MediaTypeManifest, manifest.MediaType)
	}

	if manifest.Config.Digest != config.Digest {
		t.Errorf("expected config digest %s, got %s", config.Digest, manifest.Config.Digest)
	}

	if len(manifest.Layers) != len(layers) {
		t.Errorf("expected %d layers, got %d", len(layers), len(manifest.Layers))
	}

	if manifest.Annotations == nil {
		t.Errorf("expected annotations to be initialized")
	}
}

func TestComputeDigest(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    string
		wantErr bool
	}{
		{
			name: "valid data",
			data: []byte(`{"test": "data"}`),
			want: "sha256:40b61fe1b15af0a4d5402735b26343e8cf8a045f4d81710e6108a21d91eaf366",
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ComputeDigest(tt.data)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("expected digest %s, got %s", tt.want, got)
			}
		})
	}
}

func TestFilterManifestsByPlatform(t *testing.T) {
	manifests := []Descriptor{
		{
			MediaType: MediaTypeManifest,
			Digest:    "sha256:amd64",
			Size:      1024,
			Platform: &Platform{
				Architecture: ArchitectureAmd64,
				OS:           OSLinux,
			},
		},
		{
			MediaType: MediaTypeManifest,
			Digest:    "sha256:arm64",
			Size:      1024,
			Platform: &Platform{
				Architecture: ArchitectureArm64,
				OS:           OSLinux,
			},
		},
		{
			MediaType: MediaTypeManifest,
			Digest:    "sha256:windows",
			Size:      1024,
			Platform: &Platform{
				Architecture: ArchitectureAmd64,
				OS:           OSWindows,
			},
		},
		{
			MediaType: MediaTypeManifest,
			Digest:    "sha256:no-platform",
			Size:      1024,
			Platform:  nil,
		},
	}

	tests := []struct {
		name     string
		platform *Platform
		want     int
	}{
		{
			name:     "nil platform returns all",
			platform: nil,
			want:     4,
		},
		{
			name: "filter by amd64 linux",
			platform: &Platform{
				Architecture: ArchitectureAmd64,
				OS:           OSLinux,
			},
			want: 1,
		},
		{
			name: "filter by arm64",
			platform: &Platform{
				Architecture: ArchitectureArm64,
				OS:           OSLinux,
			},
			want: 1,
		},
		{
			name: "filter by non-existent platform",
			platform: &Platform{
				Architecture: "mips",
				OS:           OSLinux,
			},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterManifestsByPlatform(manifests, tt.platform)
			if len(result) != tt.want {
				t.Errorf("expected %d manifests, got %d", tt.want, len(result))
			}
		})
	}
}

func TestGetManifestDigest(t *testing.T) {
	manifest := &Manifest{
		SchemaVersion: SchemaVersion,
		MediaType:     MediaTypeManifest,
		Config: Descriptor{
			MediaType: MediaTypeConfig,
			Digest:    "sha256:config123",
			Size:      512,
		},
		Layers: []Descriptor{
			{
				MediaType: MediaTypeLayer,
				Digest:    "sha256:layer1",
				Size:      1024,
			},
		},
	}

	digest, err := GetManifestDigest(manifest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if digest == "" {
		t.Errorf("expected non-empty digest")
	}

	if !contains(digest, "sha256:") {
		t.Errorf("expected digest to start with sha256:, got %s", digest)
	}

	// Test with invalid manifest (should still work since we just marshal)
	digest2, err := GetManifestDigest(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if digest == digest2 {
		t.Errorf("expected different digests for different manifests")
	}
}

// contains is a simple string containment check (consolidated from duplicates)
func contains(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
