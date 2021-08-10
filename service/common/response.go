package common

import (
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/infra/dataStructure/lru"
	"net/http"
)

type Code string

type Resp struct {
	Status int
	Body   gin.H
}

var RespCache = lru.New(lru.FixedLength, 10)

const (
	SUCCESS         = "SUCCESS"
	FAIL            = "FAIL"
	UNAUTHORIZED    = "UNAUTHORIZED"
	RequestIdHeader = "X-V2raya-Request-Id"
)

//当code为FAIL时，data为string类型返回给前端的消息
func Response(ctx *gin.Context, code Code, data interface{}) (status int, body gin.H) {
	if reqId := ctx.GetHeader(RequestIdHeader); reqId != "" {
		if resp := RespCache.Get(reqId); resp != nil {
			resp := resp.(Resp)
			ctx.JSON(resp.Status, resp.Body)
			return resp.Status, resp.Body
		}
		defer func() {
			RespCache.Insert(reqId, Resp{
				Status: status,
				Body:   body,
			})
		}()
	}
	status = http.StatusOK
	if code == UNAUTHORIZED {
		code = FAIL
		status = http.StatusUnauthorized
	}
	if code == FAIL {
		switch data.(type) {
		case string:
			data = data.(string)
			body = gin.H{
				"code":    code,
				"message": data,
				"data":    nil,
			}
		default:
			body = gin.H{
				"code":    code,
				"message": nil,
				"data":    data,
			}
		}
		ctx.JSON(status, body)
		return status, body
	}
	body = gin.H{
		"code":    code,
		"message": nil,
		"data":    data,
	}
	ctx.JSON(status, body)
	return status, body
}

func ResponseError(ctx *gin.Context, err error) {
	Response(ctx, FAIL, err.Error())
}
func ResponseSuccess(ctx *gin.Context, data interface{}) {
	Response(ctx, SUCCESS, data)
}
