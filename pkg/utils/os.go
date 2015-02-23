package utils

import (
	"os/exec"
	"runtime"
	"strings"
)

func IsBoot2Docker() bool {
	if runtime.GOOS == "darwin" {
		return true
	}
	return false
}

func Boot2DocketIp() (string, error) {
	cmd := exec.Command("boot2docker", "ip")
	data, err := cmd.Output()
	if err != nil {
		return "", err
	}
	addr := string(data)
	addr = strings.TrimSpace(addr)
	return addr, nil

}
