package report

import (
	"fmt"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/ipforward"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/db/configure"
	"strings"
)

func init() {
	conf.RegisterReportType(
		conf.ReportType{
			Name: "connection",
			Desc: "print the Report of current connection",
			Func: ConnectionReport,
		},
	)
}

func ConnectionReport(arg []string) (report string) {
	var lines []string
	defer func() {
		report = strings.Join(lines, "\n")
	}()

	// supplementary information
	setting := configure.GetSettingNotNil()
	lines = append(lines,
		fmt.Sprintf("IP Forward: %v/%v", ipforward.IsIpForwardOn(), setting.IpForward),
		fmt.Sprintf("Port Sharing: %v", setting.PortSharing),
		"",
	)

	// get version of v2ray-core
	ver, err := where.GetV2rayServiceVersion()
	if err != nil {
		lines = append(lines, fmt.Sprintf("failed to get version of v2ray-core: %v", err))
		return
	} else {
		lines = append(lines, fmt.Sprintf("Core Version: %v", ver))
	}
	// check if v2ray-core is running
	ok, report := DefaultCoreStatusReporter.FromDatabase()
	lines = append(lines, report)
	if !ok {
		return
	}
	ok, report = DefaultCoreStatusReporter.FromApiListening()
	lines = append(lines, report, "")
	if !ok {
		return
	}

	// check dns
	_, report = DefaultDnsReporter.DialDefaultDns()
	lines = append(lines, report)
	_, report = DefaultDnsReporter.Dial()
	lines = append(lines, report, "")

	// check preset port
	_, report = DefaultCurlReporter.PresetPortReport()
	lines = append(lines, fmt.Sprintf("%v", report))

	// check transparent proxy
	_, report = DefaultCurlReporter.TransparentReport()
	lines = append(lines, fmt.Sprintf("%v", report))
	return
}
