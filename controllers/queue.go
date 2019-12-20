package controllers

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"sync"

	"github.com/huajiao-tv/dashboard/config"
	"github.com/huajiao-tv/dashboard/crontab"
	"github.com/huajiao-tv/dashboard/keeper"
	"github.com/huajiao-tv/dashboard/models"
	"github.com/youlu-cn/ginp"
)

var Queue = new(QueueController)

type QueueController struct {
}

func (c QueueController) Group() string {
	return "queue"
}

func (c QueueController) TokenRequired(string) bool {
	return true
}

func (c QueueController) ListHandler(_ *ginp.Request) *ginp.Response {
	type QueueData struct {
		*models.Queue
		TopicCount int     `json:"topics"`
		AddQps     float64 `json:"qps"`
	}
	queues, err := models.Queue{}.Query(&models.Query{})
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	data := make([]*QueueData, 0, len(queues))
	for _, item := range queues {
		qh := crontab.QueueCollectStats.Get(item.Name)
		item.Password = "" // 抹去密码
		data = append(data, &QueueData{
			Queue:      item,
			TopicCount: item.TopicCount,
			AddQps:     math.Ceil(qh.AddQps),
		})
	}
	return ginp.DataResponse(data)
}

func (c QueueController) ListSystemHandler(_ *ginp.Request) *ginp.Response {
	type QueueSystestmData struct {
		System *models.System  `json:"system"`
		Queue  []*models.Queue `json:"queue"`
	}
	sys, err := models.System{}.Query(&models.Query{})
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	data := make([]*QueueSystestmData, 0)
	for _, sy := range sys {
		topics, err := models.Topic{}.Query(&models.Query{System: sy.Name})
		if err != nil {
			return ginp.ErrorResponse(http.StatusInternalServerError, err)
		}
		tmp := map[string]int{}
		for _, topic := range topics {
			tmp[topic.Queue] = 1
		}
		queueNames := make([]string, 0, len(tmp))
		for queueName := range tmp {
			queueNames = append(queueNames, queueName)
		}
		queues, err := models.Queue{}.FindAllByNames(queueNames)
		if err != nil {
			return ginp.ErrorResponse(http.StatusInternalServerError, err)
		}
		if len(queues) == 0 {
			continue
		}
		data = append(data, &QueueSystestmData{System: sy, Queue: queues})
	}
	blankQueue, err := models.Queue{}.FindBlankQueue()
	if err == nil && len(blankQueue) > 0 {
		data = append(data, &QueueSystestmData{System: &models.System{Name: "未使用", Describe: "空Queue"}, Queue: blankQueue})
	}
	return ginp.DataResponse(data)
}

//更新queue 只有名下topic不为空才能修改name和password
func (c QueueController) UpdateHandlerPostAction(req *ginp.Request) *ginp.Response {
	postData := struct {
		ID       uint64 `binding:"required" json:"id"`
		Describe string `binding:"required" json:"desc"`
		Comment  string `binding:"required" json:"comment"`
		Name     string ` json:"name" `
		Password string ` json:"password"`
	}{}
	if err := req.BindJSON(&postData); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}

	dbQueue, err := models.Queue{}.Get(postData.ID)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}

	if dbQueue.TopicCount > 0 {
		if postData.Name != "" && dbQueue.Name != postData.Name {
			return ginp.ErrorResponse(http.StatusBadRequest, errors.New("the queue is NOT empty"))
		}
		if postData.Password != "" && dbQueue.Password != postData.Password {
			return ginp.ErrorResponse(http.StatusBadRequest, errors.New("the queue is NOT empty"))
		}
	}

	dbQueue.Describe = postData.Describe
	dbQueue.Operator = req.GetUserInfo().Name
	dbQueue.Comment = postData.Comment

	//如果name为空 就保持原来name
	if postData.Name != "" {
		dbQueue.Name = postData.Name
	}
	//如果密码为空 就保持原来密码
	if postData.Password != "" {
		dbQueue.Password = postData.Password
	}

	if err := dbQueue.Update(); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	return ginp.DataResponse(nil)
}

func (c QueueController) AddHandlerPostAction(req *ginp.Request) *ginp.Response {
	var queue models.Queue
	if err := req.BindJSON(&queue); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	queue.Operator = req.GetUserInfo().Name
	if err := queue.Create(); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	return ginp.DataResponse(queue)
}

func (c QueueController) DeleteHandlerPostAction(req *ginp.Request) *ginp.Response {
	var form DelForm
	if err := req.Bind(&form); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	queue, err := models.Queue{}.Get(form.Id)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	if queue.TopicCount > 0 {
		return ginp.ErrorResponse(http.StatusBadRequest, errors.New("the queue is NOT empty"))
	}
	if err = queue.Delete(); err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	return ginp.DataResponse(nil)
}

func (c QueueController) ExportHandlerPostAction(req *ginp.Request) *ginp.Response {
	var exportIds struct {
		Ids []int `json:"ids"`
	}
	if err := req.Bind(&exportIds); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	// 根据id查出所有的 queue
	queues, err := models.Queue{}.FindAllByIds(exportIds.Ids)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	} else if len(queues) == 0 {
		return ginp.ErrorResponse(http.StatusBadRequest, errors.New("empty queue list"))
	}

	qCount := len(queues)
	tmpResData := make(map[string]*models.TransModelQueue) //return 的临时数据
	qNames := make([]string, qCount)
	for k, q := range queues {
		qNames[k] = q.Name
		tmpResData[q.Name] = &models.TransModelQueue{
			Name:     q.Name,
			Desc:     q.Describe,
			Password: q.Password,
			Comment:  q.Comment,
			Topics:   make([]models.TransModelTopic, 0),
		}
	}

	// 查出queue_id 查出所有的topic
	topics, err := models.Topic{}.FindAllByQueues(qNames)
	if err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	// 将topic 合并到queue
	if len(topics) > 0 {
		for _, tmp := range topics {
			tmpResData[tmp.Queue].Topics = append(tmpResData[tmp.Queue].Topics, tmp)
		}
	}

	resData := make([]models.TransModelQueue, len(queues)) //return 的最终数据
	k := 0
	for _, tmp := range tmpResData {
		resData[k] = *tmp
		k++
	}
	return ginp.DataResponse(resData)
}

// 已存在的不会被覆盖
// 依赖的system和storage如果不存在则报错
func (c QueueController) ImportHandlerPostAction(req *ginp.Request) *ginp.Response {
	postData := struct {
		Data []models.TransModelQueue `binding:"required" json:"data"`
	}{}
	if err := req.BindJSON(&postData); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}

	var (
		wg         sync.WaitGroup
		resSummary []models.ImportQueueResItem
		resMu      sync.Mutex
	)

	for _, qItem := range postData.Data {
		wg.Add(1)

		// import queue
		go func(qItem models.TransModelQueue) {
			defer wg.Done()

			res := models.ImportQueue(qItem, req.GetUserInfo().Name)

			//汇总最终结果
			resMu.Lock()
			resSummary = append(resSummary, res...)
			resMu.Unlock()
		}(qItem)
	}
	wg.Wait()

	// 如果有成功的就更新keeper
	for _, item := range resSummary {
		if item.Res == "success" {
			baseVal, err := Config.getAllQueueConfig()
			if err != nil {
				return ginp.ErrorResponse(http.StatusBadRequest, errors.New(fmt.Sprintf("get queue in keeper error: %v", err)))
			}
			tagsVal, err := Config.getTagsConfig()
			if err != nil {
				return ginp.ErrorResponse(http.StatusBadRequest, errors.New(fmt.Sprintf("get tags in keeper error: %v", err)))
			}
			comment := "import queue"
			err = keeper.ConfigClient.ManageConfig(*config.Domain, []*keeper.ConfigOperate{
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
			if err != nil {
				return ginp.ErrorResponse(http.StatusBadRequest, errors.New(fmt.Sprintf("update keeper error: %v", err)))
			}
			break
		}
	}

	return ginp.DataResponse(resSummary)
}
