package controller

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/server/service"
)

/*检查是否存在账户*/
func GetAccount(ctx *gin.Context) {
	common.ResponseSuccess(ctx, gin.H{
		"hasAnyAccounts": configure.HasAnyAccounts(),
	})
}

var loginSessions = make(chan interface{}, 1)

func PostLogin(ctx *gin.Context) {
	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	// the body is read before the one-at-a-time semaphore, so it is capped
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 4096)
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		err := badRequest("username and password", "request body must be {\"username\": string, \"password\": string}")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code": common.FAIL, "message": err.Error(), "data": nil, "errorCode": "BAD_REQUEST",
			"params": gin.H{"field": "username and password"},
		})
		return
	}
	loginSessions <- nil
	defer func() {
		time.Sleep(500 * time.Millisecond)
		<-loginSessions
	}()
	if !configure.HasAnyAccounts() {
		common.Response(ctx, common.UNAUTHORIZED, gin.H{
			"first": true,
		})
		return
	}
	jwt, err := service.Login(data.Username, data.Password)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{
		"token": jwt,
	})
}

/*注册*/
var muReg sync.Mutex

func PostAccount(ctx *gin.Context) {
	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	muReg.Lock()
	defer muReg.Unlock()
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, badRequest("username and password", "request body must be {\"username\": string, \"password\": string}"))
		return
	}
	if ok, err := service.ValidPasswordLength(data.Password); !ok {
		common.ResponseError(ctx, logError(err))
		return
	}
	if configure.HasAnyAccounts() {
		common.ResponseError(ctx, common.Coded("ACCOUNT_EXISTS", logError("an account already exists; sign in instead"), nil))
		return
	}
	token, err := service.Register(data.Username, data.Password)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{
		"token": token,
	})
}
