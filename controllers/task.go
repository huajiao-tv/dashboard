package controllers

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/huajiao-tv/dashboard/crontab"
	"github.com/huajiao-tv/dashboard/dao"
	"github.com/huajiao-tv/dashboard/models"
	"github.com/youlu-cn/ginp"
)

var Task = new(TaskController)

type TaskController struct {
}

func (c TaskController) TokenRequired(string) bool {
	return true
}
func (c TaskController) Group() string {
	return "task"
}

// TODO: check circle dependency
func (c TaskController) AddHandlerPostAction(req *ginp.Request) (res *ginp.Response) {
	task := new(dao.Task)
	err := req.BindJSON(task)

	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	task.Operator = req.GetUserInfo().Name
	task.Status = dao.TaskStop
	err = task.Create()

	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	return ginp.DataResponse(task)
}

// TODO: check circle dependency
func (c TaskController) UpdateHandlerPostAction(req *ginp.Request) (res *ginp.Response) {
	task := new(dao.Task)
	err := req.BindJSON(task)

	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	oldTask := new(dao.Task)
	oldTask.Get(int64(task.ID))
	task.CreatedAt = oldTask.CreatedAt
	task.Operator = req.GetUserInfo().Name
	task.Status = dao.TaskStop
	err = task.Update()
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}

	_ = models.NewTask().StopTask(oldTask)
	return ginp.DataResponse(task)
}

func (c TaskController) ListHandlerGetAction(_ *ginp.Request) (res *ginp.Response) {
	t := dao.Task{}
	list := t.GetAllWork()
	var ret []*TaskD
	for _, val := range list {
		ret = append(ret, &TaskD{
			Task:    val,
			Statics: models.NewTask().GetInfo(val),
		})
	}
	return ginp.DataResponse(ret)
}

// TODO 完善删除任务的和停止任务的依赖检查
func (c TaskController) RemoveHandlerPostAction(req *ginp.Request) (res *ginp.Response) {
	task := new(dao.Task)
	id, _ := ioutil.ReadAll(req.Body)
	tid, err := strconv.ParseInt(string(id), 10, 64)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	task.Get(tid)
	task.Operator = req.GetUserInfo().Name
	dependency := models.NewTask().GetDependency()
	var mgs string
	deps, ok := dependency[task.Name]
	if ok {
		mgs = task.Name + "被" + strings.Join(deps, ",") + "依赖"
	}
	task.Status = dao.TaskRemove
	err = task.Update()
	if err == nil {
		_ = models.NewTask().StopTask(task)
		return ginp.DataResponse(mgs)
	} else {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
}

func (c TaskController) StartHandlerPostAction(req *ginp.Request) (res *ginp.Response) {
	task := new(dao.Task)
	id, _ := ioutil.ReadAll(req.Body)
	tid, err := strconv.ParseInt(string(id), 10, 64)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	task.Get(tid)
	task.Operator = req.GetUserInfo().Name
	task.Status = dao.TaskWorking
	err = task.Update()
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	if err := models.NewTask().Add(task); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	return ginp.DataResponse(nil)
}

func (c TaskController) PauseHandlerPostAction(req *ginp.Request) (res *ginp.Response) {
	task := new(dao.Task)
	id, _ := ioutil.ReadAll(req.Body)
	tid, err := strconv.ParseInt(string(id), 10, 64)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	task.Get(tid)
	task.Operator = req.GetUserInfo().Name
	task.Status = dao.TaskStop
	err = task.Update()
	if err == nil {
		err = models.NewTask().StopTask(task)
		if err != nil {
			return ginp.ErrorResponse(http.StatusBadRequest, err)
		}
		return ginp.DataResponse(nil)
	} else {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
}

func (c TaskController) NodesHandlerGetAction(req *ginp.Request) (res *ginp.Response) {
	task := new(dao.Task)
	tid, err := strconv.ParseInt(req.FormValue("id"), 10, 64)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	task.Get(tid)

	machines, err := dao.Machine{Cron: true}.Query(&dao.Query{System: task.System})
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	type JobNode struct {
		Node  string            `json:"node"`
		Stats map[string]string `json:"stats"`
	}
	ret := make([]*JobNode, 0, len(machines))
	for _, m := range machines {
		// TODO: add cache
		ret = append(ret, &JobNode{
			Node:  m.IP,
			Stats: models.NewTask().GetNodeWorkInfo(task, m.IP),
		})
	}
	return ginp.DataResponse(ret)
}

func (c TaskController) TreeHandlerGetAction(req *ginp.Request) (res *ginp.Response) {
	task := new(dao.Task)
	id := req.FormValue("id")
	tid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}

	task.Get(tid)
	r := task.GetDepsTree()
	return ginp.DataResponse(r)
}

func (c TaskController) NodeLogHandlerGetAction(req *ginp.Request) (res *ginp.Response) {
	task := new(dao.Task)
	id := req.FormValue("id")
	page := req.FormValue("page")
	if page == "" {
		page = "1"
	}
	tid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	node := req.FormValue("node")
	task.Get(tid)
	pageId, err := strconv.ParseInt(page, 10, 64)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, fmt.Errorf("page id err: %w", err))
	}
	searchWord := req.FormValue("search")
	ret := dao.JobResult{}.Query(task.ID, node, pageId, "desc", 10, searchWord)
	return ginp.DataResponse(ret)
}

func (c TaskController) NodeLogTotalHandlerGetAction(req *ginp.Request) (res *ginp.Response) {
	task := new(dao.Task)
	id := req.FormValue("id")
	tid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	node := req.FormValue("node")
	searchWord := req.FormValue("search")
	task.Get(tid)
	count := dao.JobResult{}.Count(task.ID, node, searchWord)
	return ginp.DataResponse(count)
}

func (c TaskController) TaskHandlerGetAction(req *ginp.Request) (res *ginp.Response) {
	task := new(dao.Task)
	id := req.FormValue("id")
	tid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	task.Get(tid)
	return ginp.DataResponse(task)
}

func (c TaskController) TestHandlerGetAction(req *ginp.Request) (res *ginp.Response) {
	task := new(dao.Task)
	id := req.FormValue("id")
	tid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	task.Get(tid)
	if task.Status == 0 {
		return ginp.ErrorResponse(http.StatusBadRequest, errors.New("please stop the task first"))
	}
	err = models.NewTask().AddTest(task, req.GetUserInfo().Name)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	return ginp.DataResponse(task)
}

func (c TaskController) TestLogHandlerGetAction(req *ginp.Request) (res *ginp.Response) {
	task := new(dao.Task)
	id := req.FormValue("id")
	tid, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	task.Get(tid)
	testTask := dao.TaskTest{}.Get(task.ID)
	ret := dao.JobResult{}.Get(testTask.TaskJobId)
	return ginp.DataResponse(ret)
}

func (c TaskController) NodeListHandlerGetAction(_ *ginp.Request) (res *ginp.Response) {
	data := crontab.TaskState.Get()
	return ginp.DataResponse(data)
}
