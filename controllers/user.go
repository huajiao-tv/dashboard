package controllers

import (
	"net/http"
	"strings"

	"github.com/huajiao-tv/dashboard/config"
	"github.com/huajiao-tv/dashboard/crontab"
	"github.com/huajiao-tv/dashboard/dao"
	"github.com/huajiao-tv/dashboard/models"
	"github.com/pkg/errors"
	"github.com/youlu-cn/ginp"
)

var User = new(UserController)

type UserController struct {
}

func (c UserController) Group() string {
	return "auth"
}

func (c UserController) TokenRequired(path string) bool {
	if strings.HasSuffix(path, "login") {
		return false
	}
	return true
}

func (c *UserController) TestLoginHandler(_ *ginp.Request) *ginp.Response {
	t := crontab.TopicMachineMetrics.GetMetrics()
	return ginp.DataResponse(t)
}

func (c *UserController) LoginHandlerPostAction(req *ginp.Request) *ginp.Response {
	var user dao.User
	if err := req.Bind(&user); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	dbUser, err := user.Get()
	if err == dao.UserNotExist {
		if !config.GlobalConfig.LDAP.Enable {
			return ginp.ErrorResponse(http.StatusBadRequest, errors.New("username or password err"))
		}
		info, err := models.Login(user.Name, user.Password)
		if err != nil {
			return ginp.ErrorResponse(http.StatusBadRequest, err)
		}
		// create new user
		err = dao.User{
			Name:    info.Name,
			Email:   info.Email,
			Avatar:  info.Avatar,
			AuthSrc: dao.LdapUser,
		}.Create()
		if err != nil {
			return ginp.ErrorResponse(http.StatusInternalServerError, err)
		}
		//
		return ginp.TokenResponse(info, "")
	} else if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	// login
	if dbUser.AuthSrc == dao.LdapUser {
		if _, err := models.Login(user.Name, user.Password); err != nil {
			return ginp.ErrorResponse(http.StatusInternalServerError, err)
		}
	} else if dbUser.Password != user.Password {
		return ginp.ErrorResponse(http.StatusBadRequest, dao.InvalidCredential)
	}
	return ginp.TokenResponse(&ginp.UserInfo{
		Name:   dbUser.Name,
		Email:  dbUser.Email,
		Avatar: dbUser.Avatar,
		Roles:  strings.Split(dbUser.Roles, ","),
	}, "")
}

func (c UserController) InfoHandler(req *ginp.Request) *ginp.Response {
	user := req.GetUserInfo()
	if user == nil {
		return ginp.ErrorResponse(http.StatusUnauthorized, nil)
	}
	return ginp.DataResponse(user)
}

func (c UserController) ListHandler(req *ginp.Request) *ginp.Response {
	if !dao.CheckPermission(req.GetUserInfo(), dao.Administration) {
		return ginp.ErrorResponse(http.StatusForbidden, errors.New("no permission"))
	}
	v, err := dao.SearchUser(req.FormValue("keyword"))
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	return ginp.DataResponse(v)
}

func (c UserController) DelHandlerPostAction(req *ginp.Request) *ginp.Response {
	if !dao.CheckPermission(req.GetUserInfo(), dao.Administration) {
		return ginp.ErrorResponse(http.StatusForbidden, errors.New("no permission"))
	}

	u := &dao.User{}
	if err := req.BindJSON(u); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	if u.ID == 0 {
		return ginp.ErrorResponse(http.StatusBadRequest, errors.New("empty id"))
	}
	var err error
	if u, err = (*u).GetByID(); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	if u.Name == req.GetUserInfo().Name {
		return ginp.ErrorResponse(http.StatusBadRequest, errors.New("invalid operation"))
	}
	if err := u.Delete(); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	return ginp.DataResponse(u.ID)
}

func (c UserController) EditrolesHandlerPostAction(req *ginp.Request) *ginp.Response {
	if !dao.CheckPermission(req.GetUserInfo(), dao.Administration) {
		return ginp.ErrorResponse(http.StatusForbidden, errors.New("no permission"))
	}
	u := &dao.User{}
	if err := req.BindJSON(&u); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	if u.ID == 0 {
		return ginp.ErrorResponse(http.StatusBadRequest, errors.New("empty id"))
	}
	roles := u.Roles
	var err error
	if u, err = (*u).GetByID(); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	u.Roles = roles
	if err := u.Update(); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	return ginp.DataResponse(u.ID)
}
