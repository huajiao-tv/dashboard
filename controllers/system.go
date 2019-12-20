package controllers

import (
	"errors"
	"net/http"

	"github.com/huajiao-tv/dashboard/crontab"
	"github.com/huajiao-tv/dashboard/models"
	"github.com/youlu-cn/ginp"
)

var System = new(SystemController)

type SystemController struct {
}

func (c SystemController) Group() string {
	return "system"
}

func (c SystemController) TokenRequired(string) bool {
	return true
}

func (c SystemController) AddHandlerPostAction(req *ginp.Request) *ginp.Response {
	system := models.System{}
	if err := req.BindJSON(&system); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	if system.Name == models.Administration {
		return ginp.ErrorResponse(http.StatusBadRequest, errors.New("the name is reserved"))
	}
	system.Operator = req.GetUserInfo().Name
	if err := system.Create(); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	return ginp.DataResponse(system)
}

func (c SystemController) ListHandler(req *ginp.Request) *ginp.Response {
	items, err := models.System{}.Query(&models.Query{})
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	var res []*models.System
	for _, v := range items {
		if models.CheckPermission(req.GetUserInfo(), v.Name) {
			v.JobSuccessCount, v.JobFailCount = crontab.SystemJobExecStats.GetCount(v.Name)
			res = append(res, v)
		}
	}
	return ginp.DataResponse(res)
}
