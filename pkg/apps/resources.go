package apps

import (
	"embed"
	"io/fs"
)

type EmbedApp struct {
	Name string
	Path string
}

var (
	//go:embed Dockerfile nginx.conf entrypoint.sh
	GitServerFS embed.FS

	//go:embed srv/*
	EmbeddedAppsFS embed.FS

	EmbedApps = []EmbedApp{{
		Name: "backstage",
		Path: "backstage",
	}, {
		Name: "crossplane",
		Path: "crossplane",
	}}
)

func GetAppsFS() (fs.FS, error) {
	return fs.Sub(EmbeddedAppsFS, "srv")
}
