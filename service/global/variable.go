package global

import "V2RayA/model/shadowsocksr"

var Version = "debug"
var FoundNew = false
var RemoteVersion = ""
var SupportTproxy = true

var ServiceControlMode SystemServiceControlMode = GetServiceControlMode()

var SSRs shadowsocksr.SSRs
