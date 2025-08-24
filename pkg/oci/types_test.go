package oci

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDescriptor_JSON(t *testing.T) {
	tests := []struct {
		name string
		desc Descriptor
		want string
	}{
		{
			name: "basic descriptor",
			desc: Descriptor{
				MediaType: MediaTypeManifest,
				Digest:    "sha256:abcd1234",
				Size:      1234,
			},
			want: `{"mediaType":"application/vnd.oci.image.manifest.v1+json","digest":"sha256:abcd1234","size":1234}`,
		},
		{
			name: "descriptor with URLs and annotations",
			desc: Descriptor{
				MediaType:   MediaTypeLayer,
				Digest:      "sha256:efgh5678",
				Size:        5678,
				URLs:        []string{"https://example.com/layer"},
				Annotations: map[string]string{"key": "value"},
			},
			want: `{"mediaType":"application/vnd.oci.image.layer.v1.tar+gzip","digest":"sha256:efgh5678","size":5678,"urls":["https://example.com/layer"],"annotations":{"key":"value"}}`,
		},
		{
			name: "descriptor with platform",
			desc: Descriptor{
				MediaType: MediaTypeManifest,
				Digest:    "sha256:ijkl9012",
				Size:      9012,
				Platform: &Platform{
					Architecture: ArchitectureAmd64,
					OS:           OSLinux,
				},
			},
			want: `{"mediaType":"application/vnd.oci.image.manifest.v1+json","digest":"sha256:ijkl9012","size":9012,"platform":{"architecture":"amd64","os":"linux"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.desc)
			if err != nil {
				t.Fatalf("failed to marshal descriptor: %v", err)
			}

			if string(data) != tt.want {
				t.Errorf("JSON mismatch:\nwant: %s\ngot:  %s", tt.want, string(data))
			}

			// Test unmarshaling
			var desc Descriptor
			if err := json.Unmarshal(data, &desc); err != nil {
				t.Fatalf("failed to unmarshal descriptor: %v", err)
			}

			if desc.MediaType != tt.desc.MediaType {
				t.Errorf("MediaType mismatch: want %s, got %s", tt.desc.MediaType, desc.MediaType)
			}
		})
	}
}

func TestPlatform_JSON(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		want     string
	}{
		{
			name: "basic platform",
			platform: Platform{
				Architecture: ArchitectureAmd64,
				OS:           OSLinux,
			},
			want: `{"architecture":"amd64","os":"linux"}`,
		},
		{
			name: "platform with variant and version",
			platform: Platform{
				Architecture: ArchitectureArm,
				OS:           OSLinux,
				OSVersion:    "5.4.0",
				Variant:      "v7",
			},
			want: `{"architecture":"arm","os":"linux","os.version":"5.4.0","variant":"v7"}`,
		},
		{
			name: "platform with features",
			platform: Platform{
				Architecture: ArchitectureAmd64,
				OS:           OSWindows,
				OSFeatures:   []string{"win32k", "hyperv"},
			},
			want: `{"architecture":"amd64","os":"windows","os.features":["win32k","hyperv"]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.platform)
			if err != nil {
				t.Fatalf("failed to marshal platform: %v", err)
			}

			if string(data) != tt.want {
				t.Errorf("JSON mismatch:\nwant: %s\ngot:  %s", tt.want, string(data))
			}

			// Test unmarshaling
			var platform Platform
			if err := json.Unmarshal(data, &platform); err != nil {
				t.Fatalf("failed to unmarshal platform: %v", err)
			}

			if platform.Architecture != tt.platform.Architecture {
				t.Errorf("Architecture mismatch: want %s, got %s", tt.platform.Architecture, platform.Architecture)
			}
		})
	}
}

func TestManifest_JSON(t *testing.T) {
	manifest := Manifest{
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
			{
				MediaType: MediaTypeLayer,
				Digest:    "sha256:layer2",
				Size:      2048,
			},
		},
		Annotations: map[string]string{
			"test": "annotation",
		},
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("failed to marshal manifest: %v", err)
	}

	var unmarshaled Manifest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal manifest: %v", err)
	}

	if unmarshaled.SchemaVersion != manifest.SchemaVersion {
		t.Errorf("SchemaVersion mismatch: want %d, got %d", manifest.SchemaVersion, unmarshaled.SchemaVersion)
	}

	if len(unmarshaled.Layers) != len(manifest.Layers) {
		t.Errorf("Layers length mismatch: want %d, got %d", len(manifest.Layers), len(unmarshaled.Layers))
	}

	if unmarshaled.Annotations["test"] != manifest.Annotations["test"] {
		t.Errorf("Annotation mismatch: want %s, got %s", manifest.Annotations["test"], unmarshaled.Annotations["test"])
	}
}

func TestReference_String(t *testing.T) {
	tests := []struct {
		name string
		ref  Reference
		want string
	}{
		{
			name: "registry with tag",
			ref: Reference{
				Registry:   "registry.example.com",
				Repository: "myapp",
				Tag:        "v1.0.0",
			},
			want: "registry.example.com/myapp:v1.0.0",
		},
		{
			name: "registry with namespace and tag",
			ref: Reference{
				Registry:   "docker.io",
				Namespace:  "library",
				Repository: "nginx",
				Tag:        "latest",
			},
			want: "docker.io/library/nginx:latest",
		},
		{
			name: "registry with digest",
			ref: Reference{
				Registry:   "gcr.io",
				Namespace:  "project",
				Repository: "image",
				Digest:     "sha256:abcd1234",
			},
			want: "gcr.io/project/image@sha256:abcd1234",
		},
		{
			name: "registry with tag and digest",
			ref: Reference{
				Registry:   "localhost:5000",
				Repository: "test",
				Tag:        "dev",
				Digest:     "sha256:efgh5678",
			},
			want: "localhost:5000/test:dev@sha256:efgh5678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ref.String()
			if got != tt.want {
				t.Errorf("Reference.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestImageConfig_JSON(t *testing.T) {
	created := time.Now()
	config := ImageConfig{
		Created:      &created,
		Author:       "test@example.com",
		Architecture: ArchitectureAmd64,
		OS:           OSLinux,
		Config: &ImageConfigConfig{
			User:       "1000:1000",
			Entrypoint: []string{"/app/start"},
			Cmd:        []string{"--config", "/etc/app.conf"},
			Env:        []string{"PATH=/usr/bin", "USER=app"},
			WorkingDir: "/app",
			Labels: map[string]string{
				"version": "1.0.0",
			},
		},
		RootFS: &RootFS{
			Type:    "layers",
			DiffIDs: []string{"sha256:diff1", "sha256:diff2"},
		},
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("failed to marshal ImageConfig: %v", err)
	}

	var unmarshaled ImageConfig
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal ImageConfig: %v", err)
	}

	if unmarshaled.Architecture != config.Architecture {
		t.Errorf("Architecture mismatch: want %s, got %s", config.Architecture, unmarshaled.Architecture)
	}

	if unmarshaled.Config.User != config.Config.User {
		t.Errorf("Config.User mismatch: want %s, got %s", config.Config.User, unmarshaled.Config.User)
	}

	if len(unmarshaled.RootFS.DiffIDs) != len(config.RootFS.DiffIDs) {
		t.Errorf("RootFS.DiffIDs length mismatch: want %d, got %d", len(config.RootFS.DiffIDs), len(unmarshaled.RootFS.DiffIDs))
	}
}