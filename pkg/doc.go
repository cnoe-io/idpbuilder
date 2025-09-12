/*
Package pkg provides idpbuilder components for OCI registry operations and authentication types.

This package contains the core types and functionality for handling OCI (Open Container Initiative)
images, manifests, and registry authentication within the idpbuilder ecosystem.

# Architecture

The package is organized into several key areas:

  - oci: OCI image and manifest types, constants, and operations
  - auth: Registry authentication and credential management types
  - certs: Certificate management types and utilities

# OCI Types

The oci subpackage provides complete OCI specification compliance including:

  - Image configuration types with platform specifications
  - Manifest handling for single and multi-platform images
  - Content descriptors and media type definitions
  - Constants for standard OCI specifications

Example usage:

	package main

	import (
		"github.com/cnoe-io/idpbuilder/pkg/oci"
	)

	func main() {
		// Create an image configuration
		config := &oci.ImageConfig{
			Architecture: oci.ArchitectureAmd64,
			OS:           oci.OSLinux,
		}

		// Parse a manifest
		manifest, err := oci.ParseManifest(manifestBytes)
		if err != nil {
			panic(err)
		}
	}

# Authentication Types

Registry authentication is handled through standardized credential types
that support various authentication methods including basic auth, bearer tokens,
and certificate-based authentication.

This package is designed to be used as a foundational layer for higher-level
idpbuilder operations that interact with OCI registries.
*/
package pkg