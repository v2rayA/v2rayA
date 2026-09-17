package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"
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
	common.ResponseSuccess(ctx, gin.H{
		"setting":             s,
		"localGFWListVersion": localGFWListVersion,
	})
}

func PutSetting(ctx *gin.Context) {
	updatingMu.Lock()
	if updating {
		common.ResponseError(ctx, processingErr)
		updatingMu.Unlock()
		return
	}
	updating = true
	updatingMu.Unlock()
	defer func() {
		updatingMu.Lock()
		updating = false
		updatingMu.Unlock()
	}()

	// Decode over the stored setting so fields the client does not send
	// (older GUIs omit the DNS cache flags, for example) keep their value
	// instead of being persisted as zero.
	data := *configure.GetSettingNotNil()
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, badRequest("settings", fmt.Errorf("request body is not a valid settings object: %v", err)))
		return
	}
	if data.MuxOn == configure.Yes && (data.Mux < 1 || data.Mux > 1024) {
		common.ResponseError(ctx, common.Coded("MUX_RANGE", logError(fmt.Errorf("mux concurrency %d is out of range; use 1-1024", data.Mux)), map[string]interface{}{"value": data.Mux}))
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
