package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"github.com/v2rayA/v2rayA/server/service"
)

func PostConnection(ctx *gin.Context) {
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

	var which configure.Which
	err := ctx.ShouldBindJSON(&which)
	if err != nil {
		common.ResponseError(ctx, logError("bad request"))
		return
	}
	err = service.Connect(&which)
	if err != nil {
		log.Warn("PostConnection: %v", err)
		common.ResponseError(ctx, logError(fmt.Errorf("failed to connect: %w", err)))
		return
	}
	getTouch(ctx)
}

func DeleteConnection(ctx *gin.Context) {
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

	var which configure.Which
	err := ctx.ShouldBindJSON(&which)
	if err != nil {
		common.ResponseError(ctx, logError("bad request"))
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

	err := service.StartV2ray()
	if err != nil {
		common.ResponseError(ctx, logError(fmt.Errorf("failed to start v2ray-core: %w", err)))
		return
	}
	getTouch(ctx)
}

func DeleteV2ray(ctx *gin.Context) {
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

	err := service.StopV2ray()
	if err != nil {
		common.ResponseError(ctx, logError(fmt.Errorf("failed to stop v2ray-core: %w", err)))
		return
	}
	getTouch(ctx)
}
