package common

import (
	"github.com/v2rayA/v2rayA/db/configure"
	"testing"
)

func TestUrlEncoded(t *testing.T) {
	str := `试试1+就试试!`
	t.Log(UrlEncoded(str))
}

func TestFillEmpty(t *testing.T) {
	setting := &configure.Setting{
		RulePortMode:                       "1",
		ProxyModeWhenSubscribe:             "2",
		GFWListAutoUpdateMode:              "3",
		GFWListAutoUpdateIntervalHour:      4,
		SubscriptionAutoUpdateMode:         "5",
		SubscriptionAutoUpdateIntervalHour: 6,
		TcpFastOpen:                        "7",
		MuxOn:                              "8",
		Mux:                                9,
		Transparent:                        "10",
		IpForward:                          false,
		PortSharing:                        false,
		SpecialMode:                        "",
		TransparentType:                    "",
		AntiPollution:                      "",
	}
	if err := FillEmpty(setting, configure.NewSetting()); err != nil {
		t.Fatal(err)
	}
	if setting.SpecialMode != configure.NewSetting().SpecialMode {
		t.Fatal()
	}
	emptySetting := &configure.Setting{}
	if err := FillEmpty(emptySetting, configure.NewSetting()); err != nil {
		t.Fatal(err)
	}
	if *emptySetting != *configure.NewSetting() {
		t.Fatal()
	}
}
