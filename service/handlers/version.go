package handlers

import (
	"github.com/gin-gonic/gin"
	"v2rayW/tools"
)

func Version(ctx *gin.Context) {
	tools.ResponseSuccess(ctx, "1.00")
}
