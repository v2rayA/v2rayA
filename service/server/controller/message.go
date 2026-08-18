package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/pkg/server/jwt"
	"github.com/v2rayA/v2rayA/server/service"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	EnableCompression: true,
}

func WsMessage(ctx *gin.Context) {
	// Validate token from query parameter or Authorization header
	token := ctx.Query("token")
	if token == "" {
		token = ctx.Query("Authorization")
	}
	if token == "" {
		token = ctx.GetHeader("Authorization")
		if strings.HasPrefix(token, "Bearer ") {
			token = token[7:]
		}
	}
	if token == "" || !jwt.ValidateToken(token) {
		common.Response(ctx, common.UNAUTHORIZED, "unauthorized")
		return
	}
	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		logError(err)
		return
	}
	h := service.NewMessageHandler(conn)
	go h.Write()
	go h.Read()
}
