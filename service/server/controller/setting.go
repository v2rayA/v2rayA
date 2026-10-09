package controller

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset"
	"github.com/v2rayA/v2rayA/server/service"
)

func GetSetting(ctx *gin.Context) {
	s := service.GetSetting()
	var localGFWListVersion string
	t, err := asset.GetGFWListModTime()
	if err == nil {
		localGFWListVersion = t.Local().Format("2006-01-02")
	}
	var localGeositeVersion string
	t, err = asset.GetGeoSiteModTime()
	if err == nil {
		localGeositeVersion = t.Local().Format("2006-01-02")
	}
	common.ResponseSuccess(ctx, gin.H{
		"setting":             s,
		"localGFWListVersion": localGFWListVersion,
		"localGeositeVersion": localGeositeVersion,
	})
}

func PutSetting(ctx *gin.Context) {
	release, ok := beginMutation(ctx)
	if !ok {
		return
	}
	defer release()

	// Decode over the stored setting so fields the client does not send
	// (older GUIs omit the DNS cache flags, for example) keep their value
	// instead of being persisted as zero.
	data := *configure.GetSettingNotNil()
	raw, err := ctx.GetRawData()
	if err != nil {
		common.ResponseError(ctx, badRequest("settings", fmt.Errorf("cannot read the request body: %v", err)))
		return
	}
	if err = json.Unmarshal(raw, &data); err != nil {
		common.ResponseError(ctx, badRequest("settings", fmt.Errorf("request body is not a valid settings object: %v", err)))
		return
	}
	if data.MuxOn == configure.Yes && (data.Mux < 1 || data.Mux > 1024) {
		common.ResponseError(ctx, common.Coded("MUX_RANGE", logError(fmt.Errorf("mux concurrency %d is out of range; use 1-1024", data.Mux)), map[string]interface{}{"value": data.Mux}))
		return
	}
	// The mode supersedes the opt-out, so only one of them decides what is
	// stored, and which one depends on what this body actually says:
	//
	//   - it carries dnsMode: the client knows the mode and it stands.
	//   - it carries only dnsHijack: the client predates the mode, and the
	//     opt-out it moved has to take effect. Decoding over the stored
	//     setting would otherwise hand it the mode already in the database
	//     and quietly drop its choice, so the mode is cleared and the
	//     migration derives it from the opt-out this body sent.
	//   - it carries neither: the client is saving some other page and never
	//     spoke about DNS. Decoding over the stored setting already left
	//     the stored mode in place, and clearing it here would resolve that
	//     same mode through the opt-out, which mirrors it — turning a save
	//     that never mentioned DNS into a downgrade of it.
	switch {
	case gjson.GetBytes(raw, "dnsMode").Exists():
	case gjson.GetBytes(raw, "dnsHijack").Exists():
		data.DnsMode = ""
	}
	if err = configure.ValidateDnsMode(data.DnsMode); err != nil {
		common.ResponseError(ctx, badRequest("dnsMode", err))
		return
	}
	// 对 DNS 配置字段执行迁移，确保旧格式请求中的缺失字段被填充默认值
	configure.MigrateSetting(&data)
	err = service.UpdateSetting(&data)
	if err != nil {
		// UpdateSetting restores the previous setting and the core with it;
		// stopping the core here would undo that and leave the user without
		// a proxy because of one rejected field.
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, nil)
}
