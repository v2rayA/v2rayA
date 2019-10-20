package handlers

import (
	"github.com/gin-gonic/gin"
	"V2RayA/tools"
)

func Version(ctx *gin.Context) {
	tools.ResponseSuccess(ctx, "1.00")
}
