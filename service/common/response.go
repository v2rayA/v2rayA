package common

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Code string

const (
	SUCCESS      = "SUCCESS"
	FAIL         = "FAIL"
	UNAUTHORIZED = "UNAUTHORIZED"
)

//当code为FAIL时，data为string类型返回给前端的消息
func Response(ctx *gin.Context, code Code, data interface{}) {
	status := http.StatusOK
	if code == UNAUTHORIZED {
		code = FAIL
		status = http.StatusUnauthorized
	}
	if code == FAIL {
		switch data.(type) {
		case string:
			data = data.(string)
			ctx.JSON(status, gin.H{
				"code":    code,
				"message": data,
				"data":    nil,
			})
		default:
			ctx.JSON(status, gin.H{
				"code":    code,
				"message": nil,
				"data":    data,
			})
		}
		return
	}
	ctx.JSON(status, gin.H{
		"code":    code,
		"message": nil,
		"data":    data,
	})
}

func ResponseError(ctx *gin.Context, err error) {
	Response(ctx, FAIL, err.Error())
}
func ResponseSuccess(ctx *gin.Context, data interface{}) {
	Response(ctx, SUCCESS, data)
}
