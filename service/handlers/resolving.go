package handlers

import (
	"V2RayA/models/nodeData"
	"V2RayA/tools"
	"encoding/json"
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
func Resolving(ctx *gin.Context) {
	var (
		n   *nodeData.NodeData
		err error
	)
	u, ok := ctx.GetQuery("url")
	if !ok {
		tools.ResponseError(ctx, errors.New("参数有误"))
		return
	}
	n, err = tools.ResolveURL(u)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(n.Config), &m)
	tools.ResponseSuccess(ctx, gin.H{"config": m})
}
