package controller

import (
	"V2RayA/common"
	"V2RayA/persistence/configure"
	"V2RayA/service"
	"errors"
	"github.com/gin-gonic/gin"
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
		common.ResponseError(ctx, errors.New("bad request"))
		return
	}
	jwt, err := service.Login(data.Username, data.Password)
	if err != nil {
		common.ResponseError(ctx, err)
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
		common.ResponseError(ctx, errors.New("bad request"))
		return
	}
	if ok, err := service.ValidPasswordLength(data.Password); !ok {
		common.ResponseError(ctx, err)
		return
	}
	username := ctx.GetString("Name")
	if !service.IsValidAccount(username, data.Password) {
		common.ResponseError(ctx, errors.New("invalid username or password"))
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
		common.ResponseError(ctx, errors.New("bad request"))
		return
	}
	if ok, err := service.ValidPasswordLength(data.Password); !ok {
		common.ResponseError(ctx, err)
		return
	}
	if configure.HasAnyAccounts() {
		common.ResponseError(ctx, errors.New("register closed"))
		return
	}
	token, err := service.Register(data.Username, data.Password)
	if err != nil {
		common.ResponseError(ctx, err)
		return
	}
	common.ResponseSuccess(ctx, gin.H{
		"token": token,
	})
}
