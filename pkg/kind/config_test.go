package kind

import (
	"io/fs"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	defaultTemplate, err := fs.ReadFile(configFS, "resources/kind.yaml.tmpl")
	if err != nil {
		t.Fatalf("failed to load default kind template: %v", err)
	}

	customTemplate, err := fs.ReadFile(configFS, "testdata/custom-kind.yaml.tmpl")
	if err != nil {
		t.Fatalf("failed to load custom kind template: %v", err)
	}

	type test struct {
		path     string
		expected []byte
	}
	tests := []test{
		{
			path:     "",
			expected: defaultTemplate,
		},
		{
			path:     "testdata/custom-kind.yaml.tmpl",
			expected: customTemplate,
		},
	}

	for _, tc := range tests {
		out, err := loadConfig(tc.path)
		if err != nil {
			t.Errorf("failed to load kind config: %v", err)
		}
		if !reflect.DeepEqual(tc.expected, out) {
			t.Errorf("expected:\n%v\ngot:\n%v", string(tc.expected), string(out))
		}
	}
}

func TestExtraPortMappingsUtilFunc(t *testing.T) {
	type test struct {
		extraPortMappings string
		expected          []PortMapping
	}
	tests := []test{
		{
			extraPortMappings: "",
			expected:          []PortMapping(nil),
		},
		{
			extraPortMappings: "22:32222",
			expected: []PortMapping{
				{
					HostPort:      "22",
					ContainerPort: "32222",
				},
			},
		},
		{
			extraPortMappings: "11:1111,33:3333,4444:4444",
			expected: []PortMapping{
				{
					HostPort:      "11",
					ContainerPort: "1111",
				},
				{
					HostPort:      "33",
					ContainerPort: "3333",
				},
				{
					HostPort:      "4444",
					ContainerPort: "4444",
				},
			},
		},
	}

	for _, tc := range tests {
		pmOutput := parsePortMappings(tc.extraPortMappings)
		if !reflect.DeepEqual(tc.expected, pmOutput) {
			t.Errorf("expected: %v, got: %v", tc.expected, pmOutput)
		}
	}
}
