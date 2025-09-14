package oci

import (
	"time"
)

// Descriptor represents a content addressable descriptor
type Descriptor struct {
	// MediaType is the media type of the referenced content
	MediaType string `json:"mediaType"`

	// Digest is the digest of the referenced content
	Digest string `json:"digest"`

	// Size is the size in bytes of the referenced content
	Size int64 `json:"size"`

	// URLs contains the list of URLs from which the content may be downloaded
	URLs []string `json:"urls,omitempty"`

	// Annotations contains arbitrary metadata relating to the referenced content
	Annotations map[string]string `json:"annotations,omitempty"`

	// Platform describes the platform which the image in the manifest runs on
	Platform *Platform `json:"platform,omitempty"`
}

// Platform describes the platform which the image in the manifest runs on
type Platform struct {
	// Architecture field specifies the CPU architecture, for example amd64 or ppc64
	Architecture string `json:"architecture"`

	// OS specifies the operating system, for example linux or windows
	OS string `json:"os"`

	// OSVersion is an optional field specifying the operating system version
	OSVersion string `json:"os.version,omitempty"`

	// OSFeatures is an optional field specifying an array of strings,
	// each listing a required OS feature (for example on Windows win32k)
	OSFeatures []string `json:"os.features,omitempty"`

	// Variant is an optional field specifying a variant of the CPU, for
	// example v7 to specify ARMv7 when architecture is arm
	Variant string `json:"variant,omitempty"`
}

// Manifest represents the OCI image manifest
type Manifest struct {
	// SchemaVersion is the image manifest schema version
	SchemaVersion int `json:"schemaVersion"`

	// MediaType specifies the type of this document data structure
	MediaType string `json:"mediaType,omitempty"`

	// Config references a configuration object for a container
	Config Descriptor `json:"config"`

	// Layers is an ordered collection of filesystem layer change descriptors
	Layers []Descriptor `json:"layers"`

	// Annotations contains arbitrary metadata for the image manifest
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ManifestList represents the OCI image manifest list (index)
type ManifestList struct {
	// SchemaVersion is the image manifest schema version
	SchemaVersion int `json:"schemaVersion"`

	// MediaType specifies the type of this document data structure
	MediaType string `json:"mediaType,omitempty"`

	// Manifests references platform specific manifests
	Manifests []Descriptor `json:"manifests"`

	// Annotations contains arbitrary metadata for the manifest list
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ImageConfig represents the OCI image configuration
type ImageConfig struct {
	// Created is the combined date and time at which the image was created
	Created *time.Time `json:"created,omitempty"`

	// Author is the name and/or email address of the person or entity which created the image
	Author string `json:"author,omitempty"`

	// Architecture is the CPU architecture which the binaries in this image are built to run on
	Architecture string `json:"architecture"`

	// OS is the name of the operating system which the image is built to run on
	OS string `json:"os"`

	// Config defines the execution parameters which should be used as a base when running a container using the image
	Config *ImageConfigConfig `json:"config,omitempty"`

	// RootFS references the layer content addresses used by the image
	RootFS *RootFS `json:"rootfs"`

	// History describes the history of each layer
	History []History `json:"history,omitempty"`
}

// ImageConfigConfig defines the execution parameters which should be used as a base when running a container
type ImageConfigConfig struct {
	// User that will run the command(s) inside the container, also support user:group
	User string `json:"User,omitempty"`

	// ExposedPorts a set of ports to expose from a container running this image
	ExposedPorts map[string]struct{} `json:"ExposedPorts,omitempty"`

	// Env is a list of environment variables to be used in a container
	Env []string `json:"Env,omitempty"`

	// Entrypoint defines a list of arguments to use as the command to execute when the container starts
	Entrypoint []string `json:"Entrypoint,omitempty"`

	// Cmd defines the default arguments to the entrypoint of the container
	Cmd []string `json:"Cmd,omitempty"`

	// Volumes is a set of directories describing where the process is likely write data specific to a container instance
	Volumes map[string]struct{} `json:"Volumes,omitempty"`

	// WorkingDir sets the current working directory of the entrypoint process in the container
	WorkingDir string `json:"WorkingDir,omitempty"`

	// Labels contains arbitrary metadata for the container
	Labels map[string]string `json:"Labels,omitempty"`

	// StopSignal contains the system call signal that will be sent to the container to exit
	StopSignal string `json:"StopSignal,omitempty"`
}

// RootFS describes the container's root filesystem
type RootFS struct {
	// Type is the type of the rootfs, usually 'layers'
	Type string `json:"type"`

	// DiffIDs is an array of layer content hashes in order from first to last
	DiffIDs []string `json:"diff_ids"`
}

// History describes the history of a layer
type History struct {
	// Created is the combined date and time at which the layer was created
	Created *time.Time `json:"created,omitempty"`

	// CreatedBy is the command which created the layer
	CreatedBy string `json:"created_by,omitempty"`

	// Author is the author of the build point
	Author string `json:"author,omitempty"`

	// Comment is a custom message set when creating the layer
	Comment string `json:"comment,omitempty"`

	// EmptyLayer is used to mark if the history item created a filesystem diff
	EmptyLayer bool `json:"empty_layer,omitempty"`
}

// Reference represents an OCI registry reference
type Reference struct {
	// Registry is the registry hostname
	Registry string `json:"registry"`

	// Namespace is the namespace/organization
	Namespace string `json:"namespace,omitempty"`

	// Repository is the repository name
	Repository string `json:"repository"`

	// Tag is the tag name
	Tag string `json:"tag,omitempty"`

	// Digest is the content digest
	Digest string `json:"digest,omitempty"`
}

// String returns the string representation of the reference
func (r *Reference) String() string {
	result := r.Registry + "/"
	if r.Namespace != "" {
		result += r.Namespace + "/"
	}
	result += r.Repository
	
	if r.Tag != "" {
		result += ":" + r.Tag
	}
	if r.Digest != "" {
		result += "@" + r.Digest
	}
	
	return result
}