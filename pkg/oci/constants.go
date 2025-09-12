package oci

// OCI Media Types
const (
	// MediaTypeManifest represents the OCI image manifest media type
	MediaTypeManifest = "application/vnd.oci.image.manifest.v1+json"
	
	// MediaTypeManifestList represents the OCI manifest list media type
	MediaTypeManifestList = "application/vnd.oci.image.index.v1+json"
	
	// MediaTypeConfig represents the OCI image configuration media type
	MediaTypeConfig = "application/vnd.oci.image.config.v1+json"
	
	// MediaTypeLayer represents the OCI image layer media type
	MediaTypeLayer = "application/vnd.oci.image.layer.v1.tar+gzip"
	
	// MediaTypeEmptyJSON represents empty JSON layer media type
	MediaTypeEmptyJSON = "application/vnd.oci.empty.v1+json"
)

// Docker Media Types (for compatibility)
const (
	// MediaTypeDockerManifest represents the Docker manifest media type
	MediaTypeDockerManifest = "application/vnd.docker.distribution.manifest.v2+json"
	
	// MediaTypeDockerManifestList represents the Docker manifest list media type
	MediaTypeDockerManifestList = "application/vnd.docker.distribution.manifest.list.v2+json"
	
	// MediaTypeDockerConfig represents the Docker image config media type
	MediaTypeDockerConfig = "application/vnd.docker.container.image.v1+json"
	
	// MediaTypeDockerLayer represents the Docker image layer media type
	MediaTypeDockerLayer = "application/vnd.docker.image.rootfs.diff.tar.gzip"
)

// Architecture constants
const (
	// ArchitectureAmd64 represents the amd64 architecture
	ArchitectureAmd64 = "amd64"
	
	// ArchitectureArm64 represents the arm64 architecture
	ArchitectureArm64 = "arm64"
	
	// ArchitectureArm represents the arm architecture
	ArchitectureArm = "arm"
	
	// Architecture386 represents the 386 architecture
	Architecture386 = "386"
	
	// ArchitecturePpc64le represents the ppc64le architecture
	ArchitecturePpc64le = "ppc64le"
	
	// ArchitectureS390x represents the s390x architecture
	ArchitectureS390x = "s390x"
)

// OS constants
const (
	// OSLinux represents the Linux operating system
	OSLinux = "linux"
	
	// OSWindows represents the Windows operating system
	OSWindows = "windows"
	
	// OSDarwin represents the Darwin (macOS) operating system
	OSDarwin = "darwin"
	
	// OSFreebsd represents the FreeBSD operating system
	OSFreebsd = "freebsd"
)

// Annotation keys
const (
	// AnnotationCreated represents the creation timestamp annotation
	AnnotationCreated = "org.opencontainers.image.created"
	
	// AnnotationAuthors represents the authors annotation
	AnnotationAuthors = "org.opencontainers.image.authors"
	
	// AnnotationURL represents the URL annotation
	AnnotationURL = "org.opencontainers.image.url"
	
	// AnnotationDocumentation represents the documentation annotation
	AnnotationDocumentation = "org.opencontainers.image.documentation"
	
	// AnnotationSource represents the source annotation
	AnnotationSource = "org.opencontainers.image.source"
	
	// AnnotationVersion represents the version annotation
	AnnotationVersion = "org.opencontainers.image.version"
	
	// AnnotationRevision represents the revision annotation
	AnnotationRevision = "org.opencontainers.image.revision"
	
	// AnnotationVendor represents the vendor annotation
	AnnotationVendor = "org.opencontainers.image.vendor"
	
	// AnnotationLicenses represents the licenses annotation
	AnnotationLicenses = "org.opencontainers.image.licenses"
	
	// AnnotationRefName represents the reference name annotation
	AnnotationRefName = "org.opencontainers.image.ref.name"
	
	// AnnotationTitle represents the title annotation
	AnnotationTitle = "org.opencontainers.image.title"
	
	// AnnotationDescription represents the description annotation
	AnnotationDescription = "org.opencontainers.image.description"
)

// Schema version
const (
	// SchemaVersion represents the OCI manifest schema version
	SchemaVersion = 2
)