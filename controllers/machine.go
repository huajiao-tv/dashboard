package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/huajiao-tv/dashboard/config"
	"github.com/huajiao-tv/dashboard/dao"
	"github.com/huajiao-tv/dashboard/keeper"
	"github.com/huajiao-tv/peppercron/logic"
	"github.com/youlu-cn/ginp"
)

var Machine = new(MachineController)

type MachineForm struct {
	Cron bool `json:"cron" form:"cron"`
}

type MachineController struct {
}

func (c MachineController) Group() string {
	return "machine"
}

func (c MachineController) TokenRequired(path string) bool {
	return true
}

func (c MachineController) ListHandler(req *ginp.Request) *ginp.Response {
	var form MachineForm
	if err := req.BindJSON(&form); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	items, err := dao.Machine{Cron: form.Cron}.Query(&dao.Query{})
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	var res []*dao.Machine
	for _, v := range items {
		if dao.CheckPermission(req.GetUserInfo(), v.System) {
			res = append(res, v)
		}
	}
	return ginp.DataResponse(res)
}

func (c MachineController) AddHandlerPostAction(req *ginp.Request) *ginp.Response {
	var machine dao.Machine
	if err := req.BindJSON(&machine); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	machine.Operator = req.GetUserInfo().Name
	if err := machine.Create(); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	if machine.Cron {
		if err := c.updateEtcd(machine.IPs, machine.Comment); err != nil {
			return ginp.ErrorResponse(http.StatusInternalServerError, err)
		}
	} else {
		// update keeper for pepper_bus
		if err := c.updateKeeper(machine.IPs, machine.Comment); err != nil {
			return ginp.ErrorResponse(http.StatusInternalServerError, err)
		}
	}
	return ginp.DataResponse(nil)
}

func (c MachineController) DeleteHandlerPostAction(req *ginp.Request) *ginp.Response {
	var form struct {
		DelForm
		MachineForm
	}
	if err := req.Bind(&form); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	machine, err := dao.Machine{Cron: form.Cron}.Get(form.Id)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	if err = machine.Delete(); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	comment := "del " + machine.IP
	if form.Cron {
		if err := c.updateEtcd(machine.IP, comment); err != nil {
			return ginp.ErrorResponse(http.StatusInternalServerError, err)
		}
	} else {
		// update keeper for pepper_bus
		if err := c.updateKeeper(machine.IP, comment); err != nil {
			return ginp.ErrorResponse(http.StatusInternalServerError, err)
		}
	}
	return ginp.DataResponse(nil)
}

func (c MachineController) updateEtcd(ips string, comment string) error {
	ms, err := dao.Machine{Cron: true}.Query(&dao.Query{Keyword: ips})
	if err != nil {
		return err
	}
	confs := make(map[string][]string)
	for _, m := range ms {
		confs[m.IP] = append(confs[m.IP], m.System)
	}
	for node, tags := range confs {
		c := CronNodeConfig{
			Tags: strings.Join(tags, ","),
		}
		v, err := json.Marshal(c)
		if err != nil {
			return err
		}
		key := fmt.Sprintf(logic.NodeConfig, node)
		if _, err = config.ETCDClient.Put(context.TODO(), key, string(v)); err != nil {
			return err
		}
	}
	return nil
}

func (c MachineController) updateKeeper(ips string, comment string) error {
	confs, err := Config.getMachineConfig(ips)
	if err != nil {
		return err
	}
	operates := make([]*keeper.ConfigOperate, 0, len(confs)*2)
	for addr, conf := range confs {
		operates = append(operates, &keeper.ConfigOperate{
			Action:  keeper.AddConfig,
			Cluster: config.GlobalConfig.Keeper.Domain,
			File:    "queue.conf",
			Section: addr,
			Key:     "tags",
			Type:    "[]string",
			Value:   strings.Join(conf.Tags, ","),
			Comment: comment,
		}, &keeper.ConfigOperate{
			Action:  keeper.AddConfig,
			Cluster: config.GlobalConfig.Keeper.Domain,
			File:    "queue.conf",
			Section: addr,
			Key:     "shard_id",
			Type:    "int",
			Value:   strconv.FormatInt(int64(conf.SharedId), 10),
			Comment: comment,
		})
	}
	return keeper.ConfigClient.ManageConfig(config.GlobalConfig.Keeper.Domain, operates, comment)
}
