package service

import (
	"bytes"
	"encoding/json"
	"time"

	ws "github.com/gorilla/websocket"
	"github.com/huajiao-tv/dashboard/config"
	"github.com/huajiao-tv/dashboard/crontab"
	"github.com/youlu-cn/ginp"
)

type Push struct {
	*ws.Conn
	Status    uint
	Timestamp time.Time
	Hash      []byte
}

func NewWebsocket(conn *ws.Conn) {
	c := &Push{
		Conn:      conn,
		Status:    0,
		Timestamp: time.Now(),
		Hash:      crontab.TaskState.GetSum(),
	}
	go c.Run()
}

func (c *Push) recv() {
	for {
		msgType, body, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}
		switch msgType {
		case ws.PingMessage:
			c.Timestamp = time.Now()
			if err := c.Conn.WriteMessage(ws.PongMessage, []byte("pong")); err != nil {
				return
			}
		case ws.TextMessage:
			var data map[string]interface{}
			if err := json.Unmarshal(body, &data); err != nil {
				return
			}
			if c.Status == 0 {
				val, ok := data["token"]
				if !ok {
					return
				}
				if _, err := ginp.NewJWT(config.TokenSign).Parse(val.(string)); err != nil {
					return
				}
				c.Status = 1
			}
		case ws.CloseMessage:
			return
		}
	}
}

func (c *Push) Run() {
	go c.recv()

	for {
		sum := crontab.TaskState.GetSum()
		if bytes.Compare(c.Hash, sum) == 0 {
			time.Sleep(time.Second)
			continue
		}

		msg, _ := json.Marshal(map[string]string{
			"job": "update",
		})
		if err := c.Conn.WriteMessage(ws.TextMessage, msg); err != nil {
			return
		}

		c.Timestamp = time.Now()
		c.Hash = sum
		time.Sleep(time.Second)
	}
}
