package main

import (
	"github.com/huajiao-tv/dashboard/config"
	"github.com/huajiao-tv/dashboard/crontab"
	"github.com/huajiao-tv/dashboard/dao"
	"github.com/huajiao-tv/dashboard/keeper"
	"github.com/huajiao-tv/dashboard/router"
)

func init() {
	// Initialize configuration
	config.Init()

	// Initialize gokeeper client
	keeper.Init()
}

func main() {
	// Initialize database models
	dao.Init()

	crontab.Init()

	// Initialize router and serve HTTP requests
	router.Init()
	router.Serve()

	select {}
}
