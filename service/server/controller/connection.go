package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"github.com/v2rayA/v2rayA/server/service"
)

var (
	refreshAutomaticGroups = service.RefreshAutomaticGroups
	startV2ray             = service.StartV2ray
)

func PostConnection(ctx *gin.Context) {
	release, ok := beginMutation(ctx)
	if !ok {
		return
	}
	defer release()

	var which configure.NodeRef
	err := ctx.ShouldBindJSON(&which)
	if err != nil {
		common.ResponseError(ctx, badRequest("server item", "request body must be a server item with _type, id, sub and outbound"))
		return
	}
	err = service.Connect(&which)
	if err != nil {
		log.Warn("Connect request failed: %v", err)
		common.ResponseError(ctx, logError(err))
		return
	}
	getTouch(ctx)
}

func DeleteConnection(ctx *gin.Context) {
	release, ok := beginMutation(ctx)
	if !ok {
		return
	}
	defer release()

	var which configure.NodeRef
	err := ctx.ShouldBindJSON(&which)
	if err != nil {
		common.ResponseError(ctx, badRequest("server item", "request body must be a server item with _type, id, sub and outbound"))
		return
	}
	err = service.Disconnect(which, false)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	getTouch(ctx)
}

func PostV2ray(ctx *gin.Context) {
	// A manual start must build the core from a freshly checked automatic
	// membership list, rather than waiting for the group's next interval.
	if err := refreshAutomaticGroups(ctx.Request.Context(), ""); err != nil {
		common.ResponseError(ctx, logError(fmt.Errorf("failed to refresh automatic groups: %w", err)))
		return
	}
	release, ok := beginMutation(ctx)
	if !ok {
		return
	}
	defer release()

	err := startV2ray()
	if err != nil {
		common.ResponseError(ctx, logError(fmt.Errorf("failed to start v2ray-core: %w", err)))
		return
	}
	getTouch(ctx)
}

func DeleteV2ray(ctx *gin.Context) {
	release, ok := beginMutation(ctx)
	if !ok {
		return
	}
	defer release()

	err := service.StopV2ray()
	if err != nil {
		common.ResponseError(ctx, logError(fmt.Errorf("failed to stop v2ray-core: %w", err)))
		return
	}
	getTouch(ctx)
}
