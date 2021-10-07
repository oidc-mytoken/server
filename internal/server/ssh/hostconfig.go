package ssh

import (
	"fmt"

	"github.com/oidc-mytoken/server/internal/config"
)

const hostEntryTemplate = `
# Host entry for mytoken
Host %s
	HostName %s
	Port %d
	User %s
	# If you use a non-default ssh key for this entry, update the following line
	# IdentityFile ~/.ssh/your.key
`

func entryName(name string) string {
	if name == "" {
		return "mytoken"
	}
	return "mytoken-" + name
}

func CreateHostConfigEntry(sshUser, name string) string {
	return fmt.Sprintf(hostEntryTemplate, entryName(name), config.Get().Host, 2222, sshUser)
}
