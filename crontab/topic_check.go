package crontab

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/huajiao-tv/dashboard/dao"
)

const (
	TopicStatusKey = "%s#%s#%s"
)

type TopicScriptsStatusCheck struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

func newTopicScriptsStatusCheck() *TopicScriptsStatusCheck {
	return &TopicScriptsStatusCheck{}
}

func (s *TopicScriptsStatusCheck) Get(queue, topic, node string) string {
	key := fmt.Sprintf(TopicStatusKey, queue, topic, node)
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, _ := json.Marshal(s.data[key])
	return string(v)
}

func (s *TopicScriptsStatusCheck) collect() {
	topics, err := dao.Topic{}.Query(&dao.Query{})
	if err != nil {
		return
	}
	for _, topic := range topics {
		if topic.RunType == 1 {
			// HTTP
			continue
		}
		go s.Update(*topic)
	}
}

func (s *TopicScriptsStatusCheck) getMachines(topic dao.Topic) []*TopicMachineMetric {
	var vals []*TopicMachineMetric
	machines, _ := dao.Machine{}.Query(&dao.Query{System: topic.System})
	for _, machine := range machines {
		stats := TopicMachineMetrics.Get(topic.Queue, topic.Name, machine.IP)
		vals = append(vals, stats)
	}
	return vals
}

func (s *TopicScriptsStatusCheck) Update(topic dao.Topic) {
	machines := s.getMachines(topic)
	for _, machine := range machines {
		status, err := SendToCgi(&SendToCgiReq{
			Queue: topic.Queue,
			Topic: topic.Name,
			Node:  machine.Node,
		})
		if err != nil {
			continue
		}

		key := fmt.Sprintf(TopicStatusKey, topic.Queue, topic.Name, machine.Node)
		s.mu.Lock()
		s.data[key] = status
		s.mu.Unlock()
	}
}
