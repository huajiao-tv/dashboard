package models

import (
	"fmt"
	"strconv"

	"github.com/huajiao-tv/dashboard/dao"
)

type Queue struct {
}

func NewQueue() *Queue {
	return &Queue{}
}

type ImportQueueResItem struct {
	Queue string `json:"queue"`
	Topic string `json:"topic"`
	Res   string `json:"res"`
}

func (m Queue) ImportQueue(qItem dao.TransModelQueue, operator string) []ImportQueueResItem {
	var res []ImportQueueResItem
	_, err := dao.Queue{}.GetByName(qItem.Name)
	if err != nil {
		queue := &dao.Queue{
			Name:     qItem.Name,
			Describe: qItem.Desc,
			Password: qItem.Password,
			Comment:  qItem.Comment,
			Operator: operator,
		}
		if err = queue.Create(); err != nil {
			return append(res, ImportQueueResItem{Queue: qItem.Name, Res: fmt.Sprintf("create queue error: %v", err)})
		}
		res = append(res, ImportQueueResItem{Queue: qItem.Name, Res: "success"})
	} else {
		res = append(res, ImportQueueResItem{Queue: qItem.Name, Res: "already exists"})
	}

	for _, tItem := range qItem.Topics {
		rItem := m.ImportTopic(qItem, tItem, operator)
		res = append(res, rItem)
	}
	return res
}

func (m Queue) ImportTopic(qItem dao.TransModelQueue, tItem dao.TransModelTopic, operator string) ImportQueueResItem {
	var topic dao.Topic
	rItem := ImportQueueResItem{
		Queue: qItem.Name,
		Topic: tItem.Name,
	}
	// check if exists
	err := topic.FindOne(map[string]string{"queue": tItem.Queue, "name": tItem.Name})
	if err == nil {
		rItem.Res = "already exists"
		return rItem
	}

	// check system and storage
	var s dao.Storage
	err = s.FindOne(map[string]string{"system": tItem.System, "host": tItem.Storage.Host, "port": tItem.Storage.Port})
	if err != nil {
		rItem.Res = fmt.Sprintf("search storage(%s:%s:%s) error: %v", tItem.System, tItem.Storage.Host, tItem.Storage.Port, err)
		return rItem
	}

	// save to db
	topic.Name = tItem.Name
	topic.Queue = tItem.Queue
	topic.Describe = tItem.Desc
	topic.Password = tItem.Password
	topic.ConsumeFile = tItem.Consume
	topic.Comment = tItem.Comment
	topic.System = tItem.System

	topic.Operator = operator
	topic.Storage = s.ID
	status, err := strconv.ParseUint(tItem.Status, 10, 8)
	if err != nil {
		rItem.Res = fmt.Sprintf("convert status error: %v", err)
		return rItem
	}
	topic.Status = uint8(status)
	err = topic.Create()
	if err != nil {
		rItem.Res = fmt.Sprintf("save topic error: %v", err)
		return rItem
	}
	rItem.Res = "success"
	return rItem
}
