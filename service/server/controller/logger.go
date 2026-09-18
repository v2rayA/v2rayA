package controller

import (
	"bufio"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"io"
	"os"

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
		common.ResponseError(ctx, badRequest("skip", "skip must be an integer"))
		return
	}
	if config.LogFile == "" {
		// No file to read: serve the in-memory tail. When the caller's
		// offset is older than what is still held, the data starts at the
		// oldest line kept; the viewer only appends, so the gap is invisible.
		data, _ := log.ReadMemory(query.Skip)
		ctx.String(200, string(data))
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
	data, err := io.ReadAll(bufio.NewReader(f))
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	ctx.String(200, string(data))
}
