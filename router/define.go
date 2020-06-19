package router

import (
	"fmt"

	"github.com/huajiao-tv/dashboard/config"
	"github.com/huajiao-tv/dashboard/controllers"
	"github.com/youlu-cn/ginp"
)

func Init() {
	ginp.SetMode(config.Mode)
	ginp.SetTokenSignKey(config.TokenSign)

	_ = ginp.AddController(controllers.User)
	_ = ginp.AddController(controllers.Machine)
	_ = ginp.AddController(controllers.Config)
	_ = ginp.AddController(controllers.Queue)
	_ = ginp.AddController(controllers.System)
	_ = ginp.AddController(controllers.Topic)
	_ = ginp.AddController(controllers.Storage)
	_ = ginp.AddController(controllers.Task)
	_ = ginp.AddController(controllers.Api)

	//_ = ginp.AddController(controllers.Push)
	ginp.Engine.GET("/push", controllers.Push.ListHandler)
}

func Serve() {
	_ = ginp.Engine.Run(fmt.Sprintf(":%d", config.GlobalConfig.Dashboard.Listen))
}
