package controller

import (
	"V2RayA/service"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
	"log"
)

func PostLogin(ctx *gin.Context) {
	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		tools.ResponseError(ctx, errors.New("bad request"))
		return
	}
	jwt, err := service.Login(data.Username, data.Password)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, gin.H{
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
		tools.ResponseError(ctx, errors.New("bad request"))
		return
	}
	if !service.ValidPasswordLength(data.Password) {
		tools.ResponseError(ctx, errors.New("length of password should be between 5 and 32"))
		return
	}
	username := ctx.GetString("Name")
	if !service.IsValidAccount(username, data.Password) {
		tools.ResponseError(ctx, errors.New("invalid username or password"))
		return
	}
	tools.ResponseSuccess(ctx, nil)
}

/*注册*/
func PostAccount(ctx *gin.Context) {
	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		tools.ResponseError(ctx, errors.New("bad request"))
		return
	}
	if !service.ValidPasswordLength(data.Password) {
		log.Println(data)
		tools.ResponseError(ctx, errors.New("length of password should be between 5 and 32"))
		return
	}
	token, err := service.Register(data.Username, data.Password)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, gin.H{
		"token": token,
	})
}
