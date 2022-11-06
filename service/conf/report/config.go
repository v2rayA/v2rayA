package report

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/v2rayA/v2rayA/conf"
)

func init() {
	conf.RegisterReportType(
		conf.ReportType{
			Name: "config",
			Desc: "print the Report of configs",
			Func: ConfigReport,
		},
	)
}

// makeEnvKey creates the environment variable key with the opts fullId and
// prefix by joining all parts together with underscores and putting all to
// upper case.
func makeEnvKey(prefix string, fullID []string) string {
	key := strings.Join(fullID, "_")
	key = strings.Replace(key, "-", "_", -1)
	key = prefix + key
	key = strings.ToUpper(key)
	return key
}

func ConfigReport(arg []string) (report string) {

	var lines []string
	defer func() {
		report = strings.Join(lines, "\n")
	}()

	t := reflect.ValueOf(conf.GetEnvironmentConfig()).Elem()
	for i := 0; i < t.NumField(); i++ {
		tag := t.Type().Field(i).Tag
		if _, ok := tag.Lookup("ignore"); ok {
			continue
		}
		id, ok := tag.Lookup("id")
		if !ok {
			continue
		}
		desc := tag.Get("desc")
		envKey := makeEnvKey("V2RAYA_", []string{id})
		value := t.Field(i).Interface()
		lines = append(lines, fmt.Sprintf("# %v", desc))
		lines = append(lines, fmt.Sprintf("%v=%v\n", envKey, value))
	}
	return
}
