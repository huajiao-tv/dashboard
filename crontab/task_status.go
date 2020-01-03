package crontab

import (
	"crypto/sha1"
	"encoding/json"
	"strconv"
	"sync"

	"github.com/huajiao-tv/dashboard/dao"
	"github.com/huajiao-tv/dashboard/models"
)

type TaskStatusCollect struct {
	mu     sync.RWMutex
	states []*dao.TaskDetail
}

func newTaskStatusCollect() *TaskStatusCollect {
	return &TaskStatusCollect{
		states: []*dao.TaskDetail{},
	}
}

func (s *TaskStatusCollect) Get() []*dao.TaskDetail {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.states
}

func (s *TaskStatusCollect) GetSum() []byte {
	s.mu.RLock()
	states := s.states
	s.mu.RUnlock()
	data, _ := json.Marshal(states)
	sum := sha1.Sum(data)
	return sum[:]
}

func (s *TaskStatusCollect) collect() {
	tasks := dao.Task{}.GetAllWork()

	states := make([]*dao.TaskDetail, 0)
	for _, val := range tasks {
		dispatch, err := models.NewTask().GetNode(val)
		if err != nil {
			continue
		}
		if dispatch == nil && val.Status == 0 {
			val.NodeStatus = 1
			_ = val.Update()
		} else if dispatch != nil && val.Status == 1 {
			val.NodeStatus = 0
			_ = val.Update()
		}

		nodes := map[string]interface{}{}
		if dispatch["running_agents"] != nil {
			nodes = dispatch["running_agents"].(map[string]interface{})
		}
		detail := &dao.TaskDetail{
			Task:  val,
			Nodes: nodes,
		}
		for k, v := range detail.Nodes {
			statics := models.NewTask().GetNodeWorkInfo(val, k)
			(v.(map[string]interface{}))["success"], _ = strconv.ParseInt(statics["success"], 10, 0)
			(v.(map[string]interface{}))["failed"], _ = strconv.ParseInt(statics["failed"], 10, 0)
			(v.(map[string]interface{}))["day_success"], _ = strconv.ParseInt(statics["day_success"], 10, 0)
			(v.(map[string]interface{}))["day_failed"], _ = strconv.ParseInt(statics["day_failed"], 10, 0)
		}
		states = append(states, detail)
	}

	s.mu.Lock()
	s.states = states
	s.mu.Unlock()
}
