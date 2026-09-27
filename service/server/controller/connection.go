package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"github.com/v2rayA/v2rayA/server/service"
)

var (
	refreshAutomaticGroups   = service.RefreshAutomaticGroups
	startV2ray               = service.StartV2ray
	needsInitialGroupRefresh = func() bool {
		for _, name := range configure.GetOutbounds() {
			if configure.GetOutboundSetting(name).AutoAdd && configure.GetConnectedServersByOutbound(name).Len() == 0 {
				return true
			}
		}
		return false
	}
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
	// An empty automatic group needs its first members before the core starts.
	// Existing members let us start immediately while checking them in the background.
	initialRefresh := needsInitialGroupRefresh()
	if initialRefresh {
		if err := refreshAutomaticGroups(ctx.Request.Context(), ""); err != nil {
			common.ResponseError(ctx, logError(fmt.Errorf("failed to refresh automatic groups: %w", err)))
			return
		}
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
	if !initialRefresh {
		go func() {
			refreshCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			if err := refreshAutomaticGroups(refreshCtx, ""); err != nil {
				log.Warn("failed to refresh automatic groups after manual start: %v", err)
			}
		}()
	}
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
