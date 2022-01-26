package controller

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"unsafe"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
)

type getLogQuery struct {
	Skip int64 `json:"skip" form:"skip"`
}

func GetLogger(ctx *gin.Context) {
	config := conf.GetEnvironmentConfig()
	query := getLogQuery{}
	if ctx.ShouldBindQuery(&query) != nil {
		common.ResponseError(ctx, errors.New("invalid query"))
		return
	}
	if config.LogFile == "" {
		if query.Skip == 0 {
			ctx.String(200, "log printed to console, please see log in console.")
		} else {
			ctx.String(200, "")
		}
		return
	}
	f, err := os.Open(config.LogFile)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	defer f.Close()
	_, err = f.Seek(query.Skip, io.SeekStart)
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	str, err := ioutil.ReadAll(bufio.NewReader(f))
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	ctx.String(200, *(*string)(unsafe.Pointer(&str)))
}
