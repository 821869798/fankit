//go:build !windows

package admin

import "errors"

func IsAdministrator() bool {
	return false
}

func RunAsAdministrator(exeFile string, exeArg []string) error {
	return errors.New("not supported run as administrator on this platform")
}
