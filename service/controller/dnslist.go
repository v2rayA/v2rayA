package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/mzz2017/v2rayA/common"
	"github.com/mzz2017/v2rayA/db/configure"
	"net"
	"strings"
)

func GetDnsList(ctx *gin.Context) {
	common.ResponseSuccess(ctx, gin.H{"dnslist": strings.Join(configure.GetDnsListNotNil(), "\n")})
}

type dnsputdata struct {
	DnsList string `json:"dnslist"`
}

func (data *dnsputdata) Valid() bool {
	str := strings.TrimSpace(data.DnsList)
	dnss := strings.Split(str, "\n")
	for _, dns := range dnss {
		dns = strings.TrimSpace(dns)
		if len(dns) <= 0 {
			continue
		}
		if net.ParseIP(dns) == nil {
			if l, e := net.LookupHost(dns); e != nil || len(l) == 0 {
				return false
			}
		}
	}
	return true
}
func (data *dnsputdata) DeDuplicate() {
	str := strings.TrimSpace(data.DnsList)
	data.DnsList = ""
	m := make(map[string]struct{})
	dnss := strings.Split(str, "\n")
	for _, dns := range dnss {
		dns = strings.TrimSpace(dns)
		if len(dns) <= 0 {
			continue
		}
		if _, ok := m[dns]; !ok {
			data.DnsList = data.DnsList + dns + "\n"
			m[dns] = struct{}{}
		}
	}
	data.DnsList = strings.TrimRight(data.DnsList, "\n")
}

func PutDnsList(ctx *gin.Context) {
	var data dnsputdata
	err := ctx.ShouldBindJSON(&data)
	if err != nil {
		common.ResponseError(ctx, logError(nil, "bad request"))
		return
	}
	if !data.Valid() {
		common.ResponseError(ctx, logError(nil, "bad format of DoH server"))
		return
	}
	data.DeDuplicate()
	err = configure.SetDnsList(&data.DnsList)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, nil)
}
