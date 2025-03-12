# Private registry authentication

idpbuilder can be configured to use private registry authentication from the
host filesystem by using the `--registry-config` flag with the `create` command.
By default this will look for a registry config file in the default
podman and docker paths (see the help text for details). You can optionally
specify a file by doing the following:
`--registry-config=$HOME/path/to/auth.json`
