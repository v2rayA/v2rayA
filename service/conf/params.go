package conf

import (
	"reflect"
	"strings"
)

// Param describes one command-line flag of v2rayA for the documentation
// page: the flag, its environment variable, its default and the help text.
type Param struct {
	Flag    string `json:"flag"`
	Short   string `json:"short"`
	Env     string `json:"env"`
	Default string `json:"default"`
	Desc    string `json:"desc"`
}

// Parameters lists the flags in Params in declaration order, read from the
// same struct tags gonfig parses, so the documentation cannot drift from
// what the binary accepts. Fields tagged ignore:"1" are one-shot commands
// (--version, --reset-password) or hidden switches and are included with
// their description when they have one.
func Parameters() []Param {
	t := reflect.TypeOf(Params{})
	params := make([]Param, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		id := f.Tag.Get("id")
		if id == "" {
			continue
		}
		desc := f.Tag.Get("desc")
		if desc == "" && f.Tag.Get("ignore") != "" {
			continue
		}
		params = append(params, Param{
			Flag:    "--" + id,
			Short:   f.Tag.Get("short"),
			Env:     "V2RAYA_" + strings.ToUpper(strings.ReplaceAll(id, "-", "_")),
			Default: f.Tag.Get("default"),
			Desc:    desc,
		})
	}
	return params
}
