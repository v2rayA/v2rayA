package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/server/service"
)

func PostImport(ctx *gin.Context) {
	var body struct {
		URL   string      `json:"url"`
		Which interface{} `json:"which"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		common.ResponseError(ctx, logError(fmt.Sprintf("bad request: %v", err)))
		return
	}

	var which *configure.Which
	if body.Which != nil {
		b, _ := jsoniter.Marshal(body.Which)
		err := jsoniter.Unmarshal(b, &which)
		if err != nil {
			common.ResponseError(ctx, logError(fmt.Sprintf("bad request (which parse error): %v", err)))
			return
		}
	}

	err := service.Import(body.URL, which)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	getTouch(ctx)
}
