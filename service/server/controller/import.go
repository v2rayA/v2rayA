package controller

import (
	"fmt"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/server/service"
)

func PostImport(ctx *gin.Context) {
	release, ok := beginMutation(ctx)
	if !ok {
		return
	}
	defer release()
	var body struct {
		URL string `json:"url"`
		// Kind is "server" or "subscription". Empty means the client did not
		// say, and the scheme of URL decides.
		Kind  string      `json:"kind"`
		Which interface{} `json:"which"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		common.ResponseError(ctx, badRequest("import", fmt.Sprintf("request body must be a JSON object with \"url\" (string) and optional \"kind\" and \"which\": %v", err)))
		return
	}

	var which *configure.NodeRef
	if body.Which != nil {
		b, _ := jsoniter.Marshal(body.Which)
		err := jsoniter.Unmarshal(b, &which)
		if err != nil {
			common.ResponseError(ctx, badRequest("which", fmt.Sprintf("\"which\" must be an object with _type, id, sub and outbound: %v", err)))
			return
		}
	}

	var err error
	switch body.Kind {
	case "server":
		err = service.ImportServer(body.URL, which)
	case "subscription":
		err = service.ImportSubscription(body.URL)
	case "":
		err = service.Import(body.URL, which)
	default:
		common.ResponseError(ctx, common.Coded("INVALID_KIND", logError(fmt.Sprintf("kind %q is not valid; expected \"server\" or \"subscription\"", body.Kind)), map[string]interface{}{"kind": body.Kind}))
		return
	}
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	getTouch(ctx)
}
