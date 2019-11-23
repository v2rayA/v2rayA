package controller

import (
	"V2RayA/model/nodeData"
	"V2RayA/service"
	"V2RayA/tools"
	"github.com/json-iterator/go"
	"errors"
	"github.com/gin-gonic/gin"
)



/*
RequestData: {
	"url": "vmess://..."
}
RequestData: {
	"url": "ss://..."
}
*/
func GetResolving(ctx *gin.Context) {
	var (
		n   *nodeData.NodeData
		err error
	)
	u, ok := ctx.GetQuery("url")
	if !ok {
		tools.ResponseError(ctx, errors.New("参数有误"))
		return
	}
	n, err = service.ResolveURL(u)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	var m map[string]interface{}
	_ = jsoniter.Unmarshal([]byte(n.Config), &m)
	tools.ResponseSuccess(ctx, gin.H{"config": m})
}
