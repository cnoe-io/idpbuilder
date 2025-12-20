# User Documentation

This directory contains user-facing documentation and guides for using IDP Builder.

## Documents

### [Minimum Requirements](./minimum-requirements.md)

System requirements for running IDP Builder.

**Requirements:**
- **CPU:** Minimum 4 cores recommended
- **Memory:** Minimum 4 GiB RAM recommended
- **Dependencies:** Docker (only runtime dependency)

**Note:** Actual requirements depend on what components you're running in your cluster.

### [Private Registry Authentication](./private-registries.md)

Guide for configuring IDP Builder to authenticate with private container registries.

**Usage:**
```bash
# Use default registry config paths (podman/docker)
idpbuilder create --registry-config

# Specify custom registry config file
idpbuilder create --registry-config=$HOME/path/to/auth.json
```

**Default Paths:**
- Podman: `$HOME/.config/containers/auth.json`
- Docker: `$HOME/.docker/config.json`

**Use Cases:**
- Pulling images from private registries
- Working in air-gapped environments
- Using enterprise container registries

## Getting Started

For installation instructions and quick start guides, see the main [README.md](../../README.md) in the repository root.

## Additional Resources

### Examples
The [examples directory](../../examples/) contains:
- Sample configurations
- Usage patterns
- Platform CR examples
- Provider CR examples

### Official Documentation
For comprehensive documentation, visit: [https://cnoe.io/docs/idpbuilder](https://cnoe.io/docs/idpbuilder)

## Related Documentation

- [Technical Specifications](../specs/) - Architectural design documents
- [Implementation Documentation](../implementation/) - Developer docs and testing info

## Getting Help

If you need help or have questions:
1. Check the main [README.md](../../README.md)
2. Browse the [examples](../../examples/)
3. Visit the [CNOE documentation](https://cnoe.io/docs/idpbuilder)
4. Open an issue on GitHub
