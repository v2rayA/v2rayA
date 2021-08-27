package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/server/service"
	"sync"
	"time"
)

var loginSessions = make(chan interface{}, 1)

func PostLogin(ctx *gin.Context) {
	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	loginSessions <- nil
	defer func() {
		time.Sleep(500 * time.Millisecond)
		<-loginSessions
	}()
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, logError("bad request"))
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

/*修改密码*/
func PutAccount(ctx *gin.Context) {
	var data struct {
		Password    string `json:"password"`
		NewPassword string `json:"newPassword"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, logError("bad request"))
		return
	}
	if ok, err := service.ValidPasswordLength(data.Password); !ok {
		common.ResponseError(ctx, logError(err))
		return
	}
	username := ctx.GetString("Name")
	if !service.IsValidAccount(username, data.Password) {
		common.ResponseError(ctx, logError("wrong username or password"))
		return
	}
	//TODO: modify password
	common.ResponseSuccess(ctx, nil)
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
		common.ResponseError(ctx, logError("bad request"))
		return
	}
	if ok, err := service.ValidPasswordLength(data.Password); !ok {
		common.ResponseError(ctx, logError(err))
		return
	}
	if configure.HasAnyAccounts() {
		common.ResponseError(ctx, logError("register closed"))
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
