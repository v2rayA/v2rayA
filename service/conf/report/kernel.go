//go:build !windows
// +build !windows

package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/v2rayA/v2rayA/conf"
	"golang.org/x/sys/unix"
)

func getKernelRelease() (string, error) {
	uname := unix.Utsname{}
	if err := unix.Uname(&uname); err != nil {
		return string(""), err
	}

	i := 0
	for ; uname.Release[i] != 0; i++ {
	}
	return string(uname.Release[:i]), nil
}

func init() {
	conf.RegisterReportType(
		conf.ReportType{
			Name: "kernel",
			Desc: "print the Report of kernel modules",
			Func: KernelReport,
		},
	)
}

func KernelReport(arg []string) (report string) {

	var lines []string
	defer func() {
		report = strings.Join(lines, "\n")
	}()

	kernelRelease, err := getKernelRelease()
	if err != nil {
		lines = append(lines, "Failed to get the uname of the running kernel.")
		lines = append(lines, "(likely non-POSIX system, but maybe a broken kernel?)")
		return
	}

	moduleRoot := filepath.Join(
		"/lib/modules",
		kernelRelease,
	)
	lines = append(lines, fmt.Sprintf("Linux Kernel Release: %v", kernelRelease))
	if s, err := os.Stat(moduleRoot); err == nil && s.IsDir() {
		lines = append(lines, fmt.Sprintf("Module Root Dir: %v", moduleRoot))
	} else {
		lines = append(lines, "Missing kernel modules, transparent proxy may not work.")
		lines = append(lines, "If you have just upgraded the linux kernel, please reboot now.")
	}
	return
}
