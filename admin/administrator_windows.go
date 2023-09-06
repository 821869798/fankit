//go:build windows

package admin

import (
	"github.com/getlantern/elevate"
	"os"
)

func IsAdministrator() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		return false
	}

	return true
}

func RunAsAdministrator(exeFile string, exeArg []string) error {
	// 如果没有禁止提权申请及非提权申请状态以及非 Linux 系统时，则进行提权操作
	// 传入进入提权状态参数
	cmd := elevate.Command(exeFile, exeArg...)
	// 开始运行
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
