//go:build !windows

package admin

import "errors"

func IsAdministrator() bool {
	return false
}

func StartRunAdministrator(exeFile string, exeArg []string) error {
	return errors.New("not supported start run exe administrator on this platform")
}

func WaitRunAdministrator(exeFile string, exeArg []string) ([]byte, error) {
	return errors.New("not supported start run exe administrator on this platform")
}
