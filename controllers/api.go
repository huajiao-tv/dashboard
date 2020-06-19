package controllers

import (
	"github.com/huajiao-tv/dashboard/crontab"
	"github.com/youlu-cn/ginp"
	"strconv"
)

var Api = new(ApiController)

type ApiController struct {
}

func (c ApiController) Group() string {
	return "api"
}

func (c ApiController) TokenRequired(string) bool {
	return false
}

func (c ApiController) AlarmHandler(req *ginp.Request) *ginp.Response {
	filter, err := strconv.Atoi(req.FormValue("filter"))
	isFilter := false
	if err == nil && filter == 1 {
		isFilter = true
	}

	type respItem struct {
		crontab.TopicAlarmEntry
		Queue          string `json:"queue"`
		Topic          string `json:"topic"`
		CurrAlarm      int64  `json:"curr_alarm"`
		CurrAlarmRetry int64  `json:"curr_alarm_retry"`
	}

	data := map[string]*respItem{}
	crontab.TopicAlarmStats.Range(func(key string, entry *crontab.TopicAlarmEntry) {
		item := new(respItem)
		item.TopicAlarmEntry = *entry
		if item.Alarm == 0 {
			item.Alarm = crontab.AlarmTh
		}
		if item.AlarmRetry == 0 {
			item.AlarmRetry = crontab.AlarmRetryTh
		}

		history := crontab.TopicLengthStats.GetByKey(key)
		item.Queue = history.Queue
		item.Topic = history.Topic
		item.CurrAlarm = history.Length
		item.CurrAlarmRetry = history.RetryLength

		if !isFilter || (history.Length > int64(item.Alarm) || history.RetryLength > int64(item.AlarmRetry)) {
			data[key] = item
		}

	})
	return ginp.DataResponse(data)
}
