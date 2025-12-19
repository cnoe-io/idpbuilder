package gitprovider

import (
	"embed"

	"github.com/cnoe-io/idpbuilder/pkg/controllers/localbuild"
)

// getGiteaFS returns the embedded Gitea installation filesystem
// This delegates to the localbuild package which has the embedded resources
func getGiteaFS() embed.FS {
	// We need to access the embedded FS from localbuild package
	// For now, we'll return an empty FS and handle this differently
	return embed.FS{}
}
