package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"

	"github.com/huajiao-tv/dashboard/config"
	"github.com/huajiao-tv/dashboard/crontab"
	"github.com/huajiao-tv/dashboard/keeper"
	"github.com/huajiao-tv/dashboard/models"
	"github.com/huajiao-tv/dashboard/service"
	"github.com/youlu-cn/ginp"
)

var Topic = new(TopicController)

type TopicController struct {
}

func (c TopicController) Group() string {
	return "topic"
}

func (c TopicController) TokenRequired(string) bool {
	return true
}

func (c TopicController) AddHandlerPostAction(req *ginp.Request) *ginp.Response {
	topic := models.Topic{}
	if err := req.BindJSON(&topic); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	if topic.Password == "" {
		if q, err := (models.Queue{}.GetByName(topic.Queue)); err != nil {
			return ginp.ErrorResponse(http.StatusInternalServerError, err)
		} else {
			topic.Password = q.Password
		}
	}
	if err := c.checkRunType(topic); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	topic.Operator = req.GetUserInfo().Name
	if err := topic.Create(); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}

	// update keeper
	if err := c.updateKeeper(topic.Comment); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	(&service.TopicService{}).UpdateScriptStatus(topic)
	return ginp.DataResponse(nil)
}

func (c TopicController) checkRunType(topic models.Topic) error {
	switch topic.RunType {
	case 0:
		return nil
	case 1: // 表示http类型
		if matched, err := regexp.MatchString("https?://[a-z0-9A-Z.-_:%#&]*$", topic.ConsumeFile); matched && err == nil {
			return nil
		} else {
			return errors.New("invalid url")
		}
	case 2:
		return nil
	default:
		return errors.New("invalid run_type")
	}
}

func (c TopicController) UpdateHandlerPostAction(req *ginp.Request) *ginp.Response {
	topic := models.Topic{}
	if err := req.BindJSON(&topic); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}

	if err := c.checkRunType(topic); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	dbTopic, err := models.Topic{}.Get(topic.ID)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	dbTopic.Name = topic.Name
	dbTopic.ConsumeFile = topic.ConsumeFile
	dbTopic.NumOfWorkers = topic.NumOfWorkers
	dbTopic.Storage = topic.Storage
	dbTopic.Operator = req.GetUserInfo().Name
	dbTopic.Comment = topic.Comment
	dbTopic.Status = topic.Status
	dbTopic.AlarmRetry = topic.AlarmRetry
	dbTopic.Alarm = topic.Alarm
	dbTopic.RunType = topic.RunType
	dbTopic.HttpConfig = topic.HttpConfig
	dbTopic.TopicConfig = topic.TopicConfig

	if err := dbTopic.Update(); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	// update keeper
	if err := Topic.updateKeeper(topic.Comment); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	(&service.TopicService{}).UpdateScriptStatus(*dbTopic)

	return ginp.DataResponse(dbTopic)
}

func (c TopicController) DeleteHandlerPostAction(req *ginp.Request) *ginp.Response {
	var form DelForm
	if err := req.Bind(&form); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	topic, err := models.Topic{}.Get(form.Id)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	if err = topic.Delete(); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	// update keeper
	comment := fmt.Sprintf("del %v/%v", topic.Queue, topic.Name)
	if err := c.updateKeeper(comment); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	return ginp.DataResponse(nil)
}

func (c TopicController) SendHandlerPostAction(req *ginp.Request) *ginp.Response {
	cgi := crontab.SendToCgiReq{}
	if err := req.BindJSON(&cgi); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}

	resp, err := crontab.SendToCgi(&cgi)
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	return ginp.DataResponse(resp)
}

func (c TopicController) DefaultThresholdHandler(_ *ginp.Request) *ginp.Response {
	return ginp.DataResponse(map[string]int{
		"alarm":       1,
		"alarm_retry": 1,
	})
}

func (c TopicController) QueueSummaryHandler(req *ginp.Request) *ginp.Response {
	query := models.Query{}
	if err := req.Bind(&query); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	if query.Queue == "" {
		return ginp.ErrorResponse(http.StatusBadRequest, errors.New("empty queue"))
	}
	topics, err := models.Topic{}.Query(&query)
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	data := make([]*TopicSummary, 0, len(topics))
	machinesList := make(map[string][]*models.Machine)
	storages := make(map[uint64]*models.Storage)
	for _, t := range topics {
		if !models.CheckPermission(req.GetUserInfo(), t.System) {
			continue
		}
		machines := machinesList[t.System]
		var err error
		if machines == nil {
			if machines, err = (models.Machine{}.Query(&models.Query{System: t.System})); err != nil {
				return ginp.ErrorResponse(http.StatusInternalServerError, err)
			}
			machinesList[t.System] = machines
		}
		st := storages[t.Storage]
		if st == nil {
			if st, err = (models.Storage{}.Get(t.Storage)); err != nil {
				return ginp.ErrorResponse(http.StatusInternalServerError, err)
			}
			storages[t.Storage] = st
		}
		length := crontab.TopicLengthStats.Get(t.Queue, t.Name)
		data = append(data, &TopicSummary{
			Id:          t.ID,
			Queue:       t.Queue,
			Name:        t.Name,
			Describe:    t.Describe,
			File:        t.ConsumeFile,
			Workers:     t.NumOfWorkers,
			Storage:     fmt.Sprintf("%v:%v", st.Host, st.Port),
			Length:      length.Length,
			RetryLen:    length.RetryLength,
			TimeoutLen:  length.TimeoutLength,
			Machines:    c.getMachineStats(t.Queue, t.Name, machines),
			System:      t.System,
			Status:      t.Status,
			Alarm:       t.Alarm,
			AlarmRetry:  t.AlarmRetry,
			RunType:     t.RunType,
			HttpConfig:  t.HttpConfig,
			TopicConfig: t.TopicConfig,
		})
	}
	return ginp.DataResponse(data)
}

func (c TopicController) SystemSummaryHandler(req *ginp.Request) *ginp.Response {
	query := models.Query{}
	if err := req.Bind(&query); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}

	if query.System == "" {
		return ginp.ErrorResponse(http.StatusBadRequest, errors.New("empty system"))
	}
	topics, err := models.Topic{}.Query(&query)
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	machines, err := models.Machine{}.Query(&query)
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	data := make([]*TopicSummary, 0, len(topics))
	storages := make(map[uint64]*models.Storage)
	for _, t := range topics {
		st := storages[t.Storage]
		if st == nil {
			if stmp, err := (models.Storage{}.Get(t.Storage)); err != nil {
				return ginp.ErrorResponse(http.StatusInternalServerError, err)
			} else {
				st = stmp
				storages[t.Storage] = st
			}
		}
		length := crontab.TopicLengthStats.Get(t.Queue, t.Name)
		data = append(data, &TopicSummary{
			Id:          t.ID,
			Queue:       t.Queue,
			Name:        t.Name,
			Describe:    t.Describe,
			File:        t.ConsumeFile,
			Workers:     t.NumOfWorkers,
			Storage:     fmt.Sprintf("%v:%v", st.Host, st.Port),
			Length:      length.Length,
			RetryLen:    length.RetryLength,
			TimeoutLen:  length.TimeoutLength,
			Machines:    c.getMachineStats(t.Queue, t.Name, machines),
			System:      t.System,
			Status:      t.Status,
			Alarm:       t.Alarm,
			AlarmRetry:  t.AlarmRetry,
			RunType:     t.RunType,
			TopicConfig: t.TopicConfig,
		})
	}
	return ginp.DataResponse(data)
}

func (c TopicController) GetRetryErrorsHandler(req *ginp.Request) *ginp.Response {
	query := models.Query{}
	if err := req.Bind(&query); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	if query.Queue == "" || query.Topic == "" {
		return ginp.ErrorResponse(http.StatusBadRequest, errors.New("empty queue or topic"))
	}
	if query.Count == 0 {
		query.Count = 10
	} else if query.Count > 50 {
		query.Count = 50
	}
	nodes, err := keeper.GetNodeList()
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	for _, node := range nodes {
		resp, err := crontab.GetCGIError(node, query.Queue, query.Topic, query.Count)
		if err != nil {
			continue
		} else {
			if len(resp.Data) <= 0 {
				continue
			}
		}
		return ginp.DataResponse(resp.Data)
	}
	return ginp.ErrorResponse(http.StatusInternalServerError, errors.New("no servers"))
}

func (c TopicController) GetLengthHandler(req *ginp.Request) *ginp.Response {
	query := models.Query{}
	if err := req.BindJSON(&query); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	if query.Queue == "" || query.Topic == "" {
		return ginp.ErrorResponse(http.StatusBadRequest, errors.New("empty queue or topic"))
	}
	nodes, err := keeper.GetNodeList()
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	data := map[string]interface{}{}
	if len(query.Machines) > 0 {
		metrics := crontab.GetTopicMachineMetrics().GetMetric()
		for _, m := range query.Machines {
			metric := metrics.Get(query.Queue, query.Topic, m)
			data[m] = metric
		}
	}

	for _, node := range nodes {
		resp, err := crontab.GetTopicLength(node, query.Queue, query.Topic)
		if err != nil {
			continue
		}
		data["topicLen"] = resp.Data
		return ginp.DataResponse(data)
	}
	return ginp.ErrorResponse(http.StatusInternalServerError, errors.New("no servers"))
}

func (c TopicController) CleanRetryHandlerPostAction(req *ginp.Request) *ginp.Response {
	query := models.Query{}
	if err := req.BindJSON(&query); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	if query.Queue == "" || query.Topic == "" {
		return ginp.ErrorResponse(http.StatusBadRequest, errors.New("empty queue or topic"))
	}
	nodes, err := keeper.GetNodeList()
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	for _, node := range nodes {
		resp, err := crontab.CleanTopicRetry(node, query.Queue, query.Topic)
		if err != nil {
			continue
		}
		return ginp.DataResponse(resp)
	}
	return ginp.ErrorResponse(http.StatusInternalServerError, errors.New("no servers"))
}

func (c TopicController) CleanStatisticsHandlerPostAction(req *ginp.Request) *ginp.Response {
	query := models.Query{}
	if err := req.BindJSON(&query); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	if query.Queue == "" || query.Topic == "" {
		return ginp.ErrorResponse(http.StatusBadRequest, errors.New("empty queue or topic"))
	}
	nodes, err := keeper.GetNodeList()
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}

	ret := make(map[string]int64)
	suc := 0
	for _, node := range nodes {
		resp, err := crontab.CleanStatistics(node, query.Queue, query.Topic)
		if err != nil {
			ret[node] = 0
			continue
		} else {
			suc++
			ret[node] = resp
		}

	}
	if suc > 0 {
		return ginp.DataResponse("success")
	} else {
		return ginp.ErrorResponse(http.StatusInternalServerError, errors.New("no servers"))
	}
}

func (c TopicController) RecoverRetryHandlerPostAction(req *ginp.Request) *ginp.Response {
	form := struct {
		models.Query
		Count int `json:"count"`
	}{}
	if err := req.BindJSON(&form); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	if form.Queue == "" || form.Topic == "" {
		return ginp.ErrorResponse(http.StatusBadRequest, errors.New("empty queue or topic"))
	}
	nodes, err := keeper.GetNodeList()
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	for _, node := range nodes {
		resp, err := crontab.RecoverTopic(node, form.Queue, form.Topic, form.Count)
		if err != nil {
			continue
		}
		return ginp.DataResponse(resp)
	}
	return ginp.ErrorResponse(http.StatusInternalServerError, errors.New("no servers"))
}

func (c TopicController) getMachineStats(queue, topic string, machines []*models.Machine) []interface{} {
	data := make([]interface{}, 0, len(machines))
	for _, machine := range machines {
		stats := crontab.GetTopicMachineMetrics().Get(queue, topic, machine.IP)
		k := fmt.Sprintf(service.TOPIC_KEY_TEMP, topic, queue, machine.IP)
		q, ok := service.TopicScript.Load(k)
		if ok {
			body, _ := json.Marshal(q)
			stats.Question = string(body)
		}
		data = append(data, stats)
	}
	return data
}

func (c TopicController) updateKeeper(comment string) error {
	baseVal, err := Config.getAllQueueConfig()
	if err != nil {
		return err
	}
	tagsVal, err := Config.getTagsConfig()
	if err != nil {
		return err
	}
	return keeper.ConfigClient.ManageConfig(*config.Domain, []*keeper.ConfigOperate{
		{
			Action:  keeper.UpdateConfig,
			Cluster: *config.Domain,
			File:    "queue.conf",
			Section: keeper.DefaultSection,
			Key:     "tags_mapping",
			Type:    "string",
			Value:   tagsVal,
			Comment: comment,
		}, {
			Action:  keeper.UpdateConfig,
			Cluster: *config.Domain,
			File:    "queue.conf",
			Section: keeper.DefaultSection,
			Key:     "queue_config",
			Type:    "string",
			Value:   baseVal,
			Comment: comment,
		},
	}, comment)
}

func (c TopicController) SearchHandler(req *ginp.Request) *ginp.Response {
	query := models.Query{}
	if err := req.BindJSON(&query); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	ks := strings.Split(query.Keyword, "/")
	topic := "%" + ks[0] + "%"
	queue := topic
	con := "or"
	if len(ks) > 1 {
		topic = "%" + ks[1] + "%"
		con = "and"
	}
	if len(query.Topic) > 0 {
		topic = query.Topic
		con = "and"
	}
	if len(query.Queue) > 0 {
		queue = query.Queue
		con = "and"
	}

	topics, err := models.Topic{}.Search(query.System, queue, topic, con)
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	data := make([]*TopicSummary, 0, len(topics))
	machinesList := make(map[string][]*models.Machine)
	storages := make(map[uint64]*models.Storage)
	for _, t := range topics {
		if !models.CheckPermission(req.GetUserInfo(), t.System) || len(data) >= 30 {
			continue
		}
		machines := machinesList[t.System]
		var err error
		if machines == nil {
			if machines, err = (models.Machine{}.Query(&models.Query{System: t.System})); err != nil {
				return ginp.ErrorResponse(http.StatusInternalServerError, err)
			}
			machinesList[t.System] = machines
		}
		st := storages[t.Storage]
		if st == nil {
			if st, err = (models.Storage{}.Get(t.Storage)); err != nil {
				return ginp.ErrorResponse(http.StatusInternalServerError, err)
			}
			storages[t.Storage] = st
		}
		length := crontab.TopicLengthStats.Get(t.Queue, t.Name)
		data = append(data, &TopicSummary{
			Id:          t.ID,
			Queue:       t.Queue,
			Name:        t.Name,
			Describe:    t.Describe,
			File:        t.ConsumeFile,
			Workers:     t.NumOfWorkers,
			Storage:     fmt.Sprintf("%v:%v", st.Host, st.Port),
			Length:      length.Length,
			RetryLen:    length.RetryLength,
			TimeoutLen:  length.TimeoutLength,
			Machines:    c.getMachineStats(t.Queue, t.Name, machines),
			System:      t.System,
			Status:      t.Status,
			Alarm:       t.Alarm,
			AlarmRetry:  t.AlarmRetry,
			RunType:     t.RunType,
			HttpConfig:  t.HttpConfig,
			TopicConfig: t.TopicConfig,
		})
	}
	return ginp.DataResponse(data)
}

func (c TopicController) HistoryHandlerPostAction(req *ginp.Request) *ginp.Response {

	postData := struct {
		Queue string `binding:"required" json:"queue"`
		Count int    `json:"count"`
	}{}
	if err := req.BindJSON(&postData); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	if postData.Count <= 0 {
		postData.Count = 10
	}

	nodes, err := keeper.GetNodeList()
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}

	node := nodes[rand.Intn(len(nodes))]
	resp, err := crontab.GetCGITaskDebug(node, postData.Queue, postData.Count)
	if err != nil {
		log.Printf("get CGI Task Debug Data error: %v", err)
		return ginp.ErrorResponse(http.StatusBadRequest, errors.New("get CGI Task Debug Data error"))
	}
	return ginp.DataResponse(resp)
}
