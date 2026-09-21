package common_test

import (
	"testing"

	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
)

func TestUrlEncoded(t *testing.T) {
	str := `试试1+就试试!`
	t.Log(common.UrlEncoded(str))
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
		TransparentType:                    "",
	}
	if err := common.FillEmpty(setting, configure.NewSetting()); err != nil {
		t.Fatal(err)
	}
	if setting.Transparent != "10" || setting.Mux != 9 {
		t.Fatalf("FillEmpty overwrote set fields: %+v", setting)
	}
	emptySetting := &configure.Setting{}
	if err := common.FillEmpty(emptySetting, configure.NewSetting()); err != nil {
		t.Fatal(err)
	}
	// bool fields are left alone on purpose (see db/configure/migrate_dns.go)
	defaults := configure.NewSetting()
	if emptySetting.Transparent != defaults.Transparent || emptySetting.Mux != defaults.Mux || emptySetting.MuxOn != defaults.MuxOn {
		t.Fatalf("FillEmpty left defaults unset: %+v", emptySetting)
	}
}
