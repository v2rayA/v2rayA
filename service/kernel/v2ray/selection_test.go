package v2ray

import "testing"

func TestApplySelection(t *testing.T) {
	infos := []serverInfo{
		socksInfo("proxy", "a"),
		socksInfo("proxy", "b"),
		socksInfo("other", "c"),
		socksInfo("other", "d"),
	}
	selected := map[string]string{
		"proxy": infos[1].Info.ExportToURL(),
		"other": "socks5://gone:1080",
	}
	kept := applySelection(infos, func(outbound string) string { return selected[outbound] })
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
