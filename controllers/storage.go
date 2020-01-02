package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"github.com/huajiao-tv/dashboard/crontab"
	"github.com/huajiao-tv/dashboard/dao"
	"github.com/youlu-cn/ginp"
)

var Storage = new(StorageController)

type StorageController struct {
}

func (c StorageController) Group() string {
	return "storage"
}

func (c StorageController) TokenRequired(string) bool {
	return true
}

func (c StorageController) ListHandler(_ *ginp.Request) *ginp.Response {
	states := crontab.RedisState.GetAll()
	return ginp.DataResponse(states)
}

func (c StorageController) AddHandlerPostAction(req *ginp.Request) *ginp.Response {
	var storage dao.Storage
	if err := req.BindJSON(&storage); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}

	if err := c.pingRedis(storage.Host, storage.Port, storage.Password); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}

	storage.Operator = req.GetUserInfo().Name
	if err := storage.Create(); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	crontab.RedisState.Update()
	return ginp.DataResponse(storage)
}

func (c StorageController) UpdateHandlerPostAction(req *ginp.Request) *ginp.Response {
	var storage dao.Storage
	if err := req.BindJSON(&storage); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	storage.Operator = req.GetUserInfo().Name
	dbStorage, err := dao.Storage{}.Get(storage.ID)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	storage.CreatedAt = dbStorage.CreatedAt
	if err := storage.Update(); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}

	if err := c.pingRedis(storage.Host, storage.Port, storage.Password); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}

	if err := Topic.updateKeeper(storage.Comment); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	return ginp.DataResponse(storage)
}

func (c StorageController) pingRedis(addr string, port int, password string) error {
	cli := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%v:%v", addr, port),
		Password:     password,
		MaxRetries:   1,
		DialTimeout:  time.Second * 3,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
	})
	defer func() {
		_ = cli.Close()
	}()
	return cli.Ping().Err()
}
