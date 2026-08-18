package v2ray

import (
	"fmt"
	"strconv"

	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

type serverInfo struct {
	Info         serverObj.ServerObj
	OutboundName string
	PluginPort   int
}

func GroupWrapper(ps string) string {
	return fmt.Sprintf("『%v』", ps)
}

type ServerData struct {
	RawServerInfos          []serverInfo
	ServerInfos             []serverInfo
	OutboundName2Setting    map[string]configure.OutboundSetting
	Link2ServerInfos        map[string][]*serverInfo
	Link2ServerObj          map[string]serverObj.ServerObj
	OutboundName2ServerObjs map[string][]serverObj.ServerObj
}

func NewServerData(serverInfos []serverInfo) (serverData *ServerData) {
	// guarantee that an v2ray outbound is reusable for balancers
	var rawServerInfos = make([]serverInfo, len(serverInfos))
	copy(rawServerInfos, serverInfos)
	link2ServerInfos := make(map[string][]*serverInfo)
	link2ServerObj := make(map[string]serverObj.ServerObj)
	for i, info := range serverInfos {
		link := info.Info.ExportToURL()
		link2ServerObj[link] = info.Info
		link2ServerInfos[link] = append(link2ServerInfos[link], &serverInfos[i])
	}
	// make ps unique
	link2ServerInfosAfter := make(map[string][]*serverInfo)
	mPsRenaming := make(map[string]struct{})
	for link, ois := range link2ServerInfos {
		ps := link2ServerObj[link].GetName()
		cnt := 2
		for {
			if _, ok := mPsRenaming[ps]; !ok {
				mPsRenaming[ps] = struct{}{}
				link2ServerObj[link].SetName(ps)
				link2ServerInfosAfter[link] = ois
				break
			}
			ps = fmt.Sprintf("%v(%v)", link2ServerObj[link].GetName(), strconv.Itoa(cnt))
			cnt++
		}
	}

	outboundName2ServerObjs := make(map[string][]serverObj.ServerObj)
	for link, ois := range link2ServerInfosAfter {
		for _, oi := range ois {
			outboundName2ServerObjs[oi.OutboundName] = append(outboundName2ServerObjs[oi.OutboundName], link2ServerObj[link])
		}
	}

	OutboundName2Setting := make(map[string]configure.OutboundSetting)
	for outbound := range outboundName2ServerObjs {
		OutboundName2Setting[outbound] = configure.GetOutboundSetting(outbound)
	}

	return &ServerData{
		RawServerInfos:          rawServerInfos,
		ServerInfos:             serverInfos,
		Link2ServerInfos:        link2ServerInfosAfter,
		Link2ServerObj:          link2ServerObj,
		OutboundName2ServerObjs: outboundName2ServerObjs,
		OutboundName2Setting:    OutboundName2Setting,
	}
}

func (sd *ServerData) ServerObj2ServerInfos() map[serverObj.ServerObj][]*serverInfo {
	m := make(map[serverObj.ServerObj][]*serverInfo)
	for link, sObj := range sd.Link2ServerObj {
		m[sObj] = sd.Link2ServerInfos[link]
	}
	return m
}

func (sd *ServerData) Ps2OutboundNames() map[string][]string {
	ps2OutboundNames := make(map[string][]string)
	for outboundName, objs := range sd.OutboundName2ServerObjs {
		for _, vi := range objs {
			ps2OutboundNames[vi.GetName()] = append(ps2OutboundNames[vi.GetName()], outboundName)
		}
	}
	return ps2OutboundNames
}

// resolveEffectiveBackend returns the effective backend ("v2ray" or "") for a ServerObj.
// It checks the node's own backend setting first, then falls back to the system setting.
