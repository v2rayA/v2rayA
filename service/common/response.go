package common

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	lru "github.com/hashicorp/golang-lru/v2"
)

type Code string

type Resp struct {
	Status int
	Body   gin.H
}

type codedResponse struct {
	err   error
	coded *CodedError
}

func codedErrorBody(err error, coded *CodedError) gin.H {
	body := gin.H{
		"code":      FAIL,
		"message":   err.Error(),
		"data":      nil,
		"errorCode": coded.Code,
	}
	if coded.Params != nil {
		body["params"] = coded.Params
	}
	return body
}

var RespCache, _ = lru.New[string, Resp](10)

const (
	SUCCESS         = "SUCCESS"
	FAIL            = "FAIL"
	UNAUTHORIZED    = "UNAUTHORIZED"
	RequestIdHeader = "X-V2raya-Request-Id"
)

// 当code为FAIL时，data为string类型返回给前端的消息
func Response(ctx *gin.Context, code Code, data interface{}) (status int, body gin.H) {
	if reqId := ctx.GetHeader(RequestIdHeader); reqId != "" {
		if resp, ok := RespCache.Get(reqId); ok {
			ctx.JSON(resp.Status, resp.Body)
			return resp.Status, resp.Body
		}
		defer func() {
			RespCache.Add(reqId, Resp{
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
		switch data := data.(type) {
		case string:
			body = gin.H{
				"code":    code,
				"message": data,
				"data":    nil,
			}
		case *CodedError:
			body = codedErrorBody(data, data)
		case codedResponse:
			body = codedErrorBody(data.err, data.coded)
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
	var coded *CodedError
	if errors.As(err, &coded) {
		Response(ctx, FAIL, codedResponse{err: err, coded: coded})
		return
	}
	Response(ctx, FAIL, err.Error())
}
func ResponseSuccess(ctx *gin.Context, data interface{}) {
	Response(ctx, SUCCESS, data)
}
