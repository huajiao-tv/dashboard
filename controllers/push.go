package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
	"github.com/huajiao-tv/dashboard/service"
)

var Push = new(PushController)

type PushController struct {
}

func (c PushController) TokenRequired(string) bool {
	return false
}

func (c PushController) Group() string {
	return "push"
}

func (c PushController) ListHandler(ctx *gin.Context) {
	updater := ws.Upgrader{
		HandshakeTimeout: 60 * time.Second,
		ReadBufferSize:   1 << 20,
		WriteBufferSize:  1 << 20,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		EnableCompression: false,
	}

	conn, err := updater.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}
	service.NewWebsocket(conn)
}
