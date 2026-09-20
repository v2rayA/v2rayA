package reqCache

import (
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"sync"
)

type requestLock struct {
	sync.Mutex
	users int
}

type reqGlen struct {
	reqIdMu map[string]*requestLock
	reqMu   sync.Mutex
}

func newReqGlen() *reqGlen {
	return &reqGlen{
		reqIdMu: make(map[string]*requestLock),
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
	if resp, ok := common.RespCache.Get(reqId); ok {
		glen.reqMu.Unlock()
		ctx.AbortWithStatusJSON(resp.Status, resp.Body)
		return
	}
	mu, ok := glen.reqIdMu[reqId]
	if !ok {
		mu = new(requestLock)
		glen.reqIdMu[reqId] = mu
	}
	mu.users++
	glen.reqMu.Unlock()
	mu.Lock()
	defer func() {
		mu.Unlock()
		glen.reqMu.Lock()
		mu.users--
		if mu.users == 0 {
			delete(glen.reqIdMu, reqId)
		}
		glen.reqMu.Unlock()
	}()
	if resp, ok := common.RespCache.Get(reqId); ok {
		ctx.AbortWithStatusJSON(resp.Status, resp.Body)
		return
	}
	ctx.Next()
}
