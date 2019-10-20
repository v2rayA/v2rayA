package handlers

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"strings"
	"V2RayA/models"
	"V2RayA/tools"
)

/*
RequestData: {
	"url": "vmess://..."
}
RequestData: {
	"url": "ss://..."
}
*/
func Resolving(ctx *gin.Context) {
	var (
		n   *models.NodeData
		err error
	)
	u, ok := ctx.GetQuery("url")
	if !ok {
		tools.ResponseError(ctx, errors.New("参数有误"))
		return
	}
	if strings.HasPrefix(u, "vmess://") {
		n, err = tools.ResolveVmessURL(u)
	} else if strings.HasPrefix(u, "ss://") {
		n, err = tools.ResolveSSURL(u)
	} else {
		tools.ResponseError(ctx, errors.New("不支持的协议"))
		return
	}
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	var m map[string]interface{}
	json.Unmarshal([]byte(n.Config), &m)
	tools.ResponseSuccess(ctx, gin.H{"config": m})
}
