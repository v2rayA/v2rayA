package reqCache

import (
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"sync"
)

type reqGlen struct {
	reqIdMu map[string]*sync.Mutex
	reqMu   sync.Mutex
}

func newReqGlen() *reqGlen {
	return &reqGlen{
		reqIdMu: make(map[string]*sync.Mutex),
		reqMu:   sync.Mutex{},
	}
}

var (
	glen = newReqGlen()
)

func ReqCache(ctx *gin.Context) {
	reqId := ctx.GetHeader(common.RequestIdHeader)
	if reqId == "" {
		return
	}
	glen.reqMu.Lock()
	if resp := common.RespCache.Get(reqId); resp != nil {
		glen.reqMu.Unlock()
		resp := resp.(common.Resp)
		ctx.AbortWithStatusJSON(resp.Status, resp.Body)
		return
	}
	mu, ok := glen.reqIdMu[reqId]
	if !ok {
		mu = new(sync.Mutex)
		glen.reqIdMu[reqId] = mu
	}
	glen.reqMu.Unlock()
	mu.Lock()
	defer mu.Unlock()
	if resp := common.RespCache.Get(reqId); resp != nil {
		resp := resp.(common.Resp)
		ctx.AbortWithStatusJSON(resp.Status, resp.Body)
		return
	}
	ctx.Next()
}
