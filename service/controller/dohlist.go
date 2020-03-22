package controller

import (
	"V2RayA/persistence/configure"
	"V2RayA/common"
	"errors"
	"github.com/gin-gonic/gin"
	"strings"
)

func GetDohList(ctx *gin.Context) {
	common.ResponseSuccess(ctx, gin.H{"dohlist": configure.GetDohListNotNil()})
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
		common.ResponseError(ctx, errors.New("bad request"))
		return
	}
	if !data.Valid() {
		common.ResponseError(ctx, errors.New("bad format of DoH server"))
		return
	}
	data.DeDuplicate()
	err = configure.SetDohList(&data.DohList)
	if err != nil {
		common.ResponseError(ctx, err)
		return
	}
	common.ResponseSuccess(ctx, nil)
}
