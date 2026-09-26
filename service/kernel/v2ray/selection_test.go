package v2ray

import (
	"testing"

	"github.com/v2rayA/v2rayA/db/configure"
)

func TestApplySelection(t *testing.T) {
	infos := []serverInfo{
		socksInfo("proxy", "a"),
		socksInfo("proxy", "b"),
		socksInfo("other", "c"),
		socksInfo("other", "d"),
	}
	settings := map[string]configure.OutboundSetting{
		"proxy": {Selected: infos[1].Info.ExportToURL()},
		"other": {Selected: "socks5://gone:1080"},
	}
	kept := applySelection(infos, func(outbound string) configure.OutboundSetting { return settings[outbound] })
	var names []string
	for _, info := range kept {
		names = append(names, info.Info.GetName())
	}
	// proxy keeps its selected member alone; other's selection matches no
	// member, so the group stays balanced
	want := []string{"b", "c", "d"}
	if len(names) != len(want) {
		t.Fatalf("kept %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("kept %v, want %v", names, want)
		}
	}
}

func TestApplySelectionKeepsStickyCurrentAndFailsClosedWhenItDisappears(t *testing.T) {
	infos := []serverInfo{socksInfo("proxy", "a"), socksInfo("proxy", "b"), socksInfo("other", "c")}
	settings := map[string]configure.OutboundSetting{
		"proxy": {
			Type:          configure.KeepCurrent,
			StickyCurrent: configure.NodeFingerprint(infos[1].Info.ExportToURL()),
		},
		"other": {Type: configure.LeastPing},
	}
	settingOf := func(outbound string) configure.OutboundSetting { return settings[outbound] }
	kept := applySelection(append([]serverInfo(nil), infos...), settingOf)
	if len(kept) != 2 || kept[0].Info.GetName() != "b" || kept[1].Info.GetName() != "c" {
		t.Fatalf("sticky selection kept %+v; want b and unrelated c", kept)
	}

	settings["proxy"] = configure.OutboundSetting{Type: configure.KeepCurrent, StickyCurrent: "missing"}
	kept = applySelection(append([]serverInfo(nil), infos...), settingOf)
	if len(kept) != 1 || kept[0].Info.GetName() != "c" {
		t.Fatalf("missing sticky selection did not fail closed: %+v", kept)
	}
}
