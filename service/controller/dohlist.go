package controller

import (
	"V2RayA/persistence/configure"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
	"strings"
)

func GetDohList(ctx *gin.Context) {
	tools.ResponseSuccess(ctx, gin.H{"dohlist": configure.GetDohListNotNil()})
}

type dohputdata struct {
	DohList string `json:"dohlist"`
}

func (data *dohputdata) Valid() bool {
	str := strings.TrimSpace(data.DohList)
	dohs := strings.Split(str, "\n")
	for _, doh := range dohs {
		doh = strings.TrimSpace(doh)
		if len(doh) <= 0 {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(doh), "https://") {
			return false
		}
	}
	return true
}
func (data *dohputdata) DeDuplicate() {
	str := strings.TrimSpace(data.DohList)
	data.DohList = ""
	m := make(map[string]struct{})
	dohs := strings.Split(str, "\n")
	for _, doh := range dohs {
		doh = strings.TrimSpace(doh)
		if len(doh) <= 0 {
			continue
		}
		if _, ok := m[doh]; !ok {
			data.DohList = data.DohList + doh + "\n"
			m[doh] = struct{}{}
		}
	}
	data.DohList = strings.TrimRight(data.DohList, "\n")
}

func PutDohList(ctx *gin.Context) {
	var data dohputdata
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		tools.ResponseError(ctx, errors.New("参数有误"))
		return
	}
	if !data.Valid() {
		tools.ResponseError(ctx, errors.New("包含无效的DoH服务器格式"))
		return
	}
	data.DeDuplicate()
	err = configure.SetDohList(&data.DohList)
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	tools.ResponseSuccess(ctx, nil)
}
