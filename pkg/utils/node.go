package utils

import (
	"os/exec"
	"strings"
)

func GetHostname() (string, error) {
	// Note: We use exec here instead of os.Hostname() because we
	// want the FQDN, and this is the easiest way to get it.
	fqdn, err := exec.Command("hostname", "-f").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(fqdn)), nil
}
