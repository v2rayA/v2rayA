// Modified from https://github.com/fatedier/frp/blob/master/pkg/util/log/log.go
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"os"

	"github.com/v2rayA/beego/v2/logs"
)

// Log is the under log object
var Log *logs.BeeLogger

func init() {
	Log = logs.NewLogger(200)
	Log.EnableFuncCallDepth(true)
	Log.SetLogFuncCallDepth(Log.GetLogFuncCallDepth() + 1)
}

func InitLog(logWay string, logFile string, logLevel string, maxdays int64, disableLogColor bool, disableTimestamp bool) {
	SetLogFile(logWay, logFile, maxdays, disableLogColor, disableTimestamp)
	SetLogLevel(logLevel)
}

// SetLogFile to configure log params
// logWay: file or console
func SetLogFile(logWay string, logFile string, maxdays int64, disableLogColor bool, disableTimestamp bool) {
	if logWay == "console" {
		params := ""
		b, _ := jsoniter.Marshal(map[string]interface{}{
			"color":     !disableLogColor,
			"timestamp": !disableTimestamp,
		})
		params = string(b)
		Log.SetLogger("console", params)
	} else {
		params := fmt.Sprintf(`{"filename": "%s", "maxdays": %d}`, logFile, maxdays)
		Log.SetLogger("file", params)
	}
}
func ParseLevel(logLevel string) int {
	level := 4 // warning
	switch logLevel {
	case "error":
		level = 3
	case "warn":
		level = 4
	case "info":
		level = 6
	case "debug":
		level = 7
	case "trace":
		level = 8
	default:
		level = 4
	}
	return level
}

// SetLogLevel set log level, default is warning
// value: error, warning, info, debug, trace
func SetLogLevel(logLevel string) {
	level := ParseLevel(logLevel)
	Log.SetLevel(level)
}

// wrap log

func Alert(format string, v ...interface{}) {
	Log.Alert(format, v...)
}

func Error(format string, v ...interface{}) {
	Log.Error(format, v...)
}

func Fatal(format string, v ...interface{}) {
	Log.Error(format, v...)
	os.Exit(1)
}

func Warn(format string, v ...interface{}) {
	Log.Warn(format, v...)
}

func Info(format string, v ...interface{}) {
	Log.Info(format, v...)
}

func Debug(format string, v ...interface{}) {
	Log.Debug(format, v...)
}

func Trace(format string, v ...interface{}) {
	Log.Trace(format, v...)
}
