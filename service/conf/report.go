package conf

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common"
	log2 "log"
	"runtime"
	"strings"
)

type reportFunc func(arg []string) (report string)

type ReportType struct {
	Name string
	Desc string
	Func reportFunc
}

var types = map[string]ReportType{}

func RegisterReportType(typ ReportType) {
	if _, ok := types[typ.Name]; ok {
		log2.Fatal("RegisterReportType: failed:", typ.Name, "exists.")
	}
	types[typ.Name] = typ
}

func PrintSupportedReports() {
	fmt.Print("The types are:\n\n")
	var lines []string
	var maxNameLength int
	for _, typ := range types {
		maxNameLength = common.Max(len(typ.Name), maxNameLength)
	}
	for _, typ := range types {
		line := "\t" + fmt.Sprintf("%"+fmt.Sprintf("%d", maxNameLength)+"s\t%s", typ.Name, typ.Desc)
		lines = append(lines, line)
	}
	fmt.Println(strings.Join(lines, "\n") + "\n")
	fmt.Println(`Use "v2raya --report <type>" to print a report.`)
}

func (p *Params) Report() {
	var arg string
	if len(p.PrintReport) > 0 {
		arg = p.PrintReport
	} else {
		return
	}
	var typName string
	fields := strings.Fields(arg)
	if len(fields) > 0 {
		typName = fields[0]
	}
	for _, typ := range types {
		if typ.Name == typName {
			report := strings.Join(append([]string{},
				fmt.Sprintf("OS: %v", runtime.GOOS),
				fmt.Sprintf("ARCH: %v", runtime.GOARCH),
				fmt.Sprintf("Go: %v", runtime.Version()),
				fmt.Sprintf("Version: %v", Version),
				"",
				typ.Func(fields[1:]),
			), "\n")
			fmt.Println(report)
			return
		}
	}
	PrintSupportedReports()
}
