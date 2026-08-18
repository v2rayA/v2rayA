// Package main is the entry point for v2raya-core, a merged binary combining
// xray-core features with v2ray-compatible MultiObservatory support.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
package main

import (
"os"

"github.com/xtls/xray-core/main/commands/base"
_ "github.com/v2rayA/v2raya-core/main/distro/all"
)

func main() {
os.Args = getArgsV4Compatible()

base.RootCommand.Long = "v2raya-core: merged xray-core with MultiObservatory support."
base.RootCommand.Commands = append(
[]*base.Command{
cmdRun,
cmdVersion,
},
base.RootCommand.Commands...,
)
base.Execute()
}

func getArgsV4Compatible() []string {
if len(os.Args) == 1 {
return []string{os.Args[0], "run"}
}
if os.Args[1][0] != '-' {
return os.Args
}
version := false
helpFlag := false
for _, a := range os.Args[1:] {
switch a {
case "-version", "--version":
version = true
case "-h", "--help":
helpFlag = true
}
}
if version {
return []string{os.Args[0], "version"}
}
if helpFlag {
return []string{os.Args[0], "help"}
}
return append([]string{os.Args[0], "run"}, os.Args[1:]...)
}
