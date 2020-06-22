package crontab

import (
	"log"
	"sync"

	"github.com/huajiao-tv/dashboard/dao"
)

var AlarmTh = 200      // 报警默认阈值
var AlarmRetryTh = 100 // retry的报警默认阈值

type TopicAlarmEntry struct {
	Alarm      int `json:"alarm"`
	AlarmRetry int `json:"alarm_retry"`
}

type TopicAlarmCollect struct {
	mu sync.RWMutex
	m  map[string]*TopicAlarmEntry
}

func newTopicAlarmCollect() *TopicAlarmCollect {
	return &TopicAlarmCollect{
		m: map[string]*TopicAlarmEntry{},
	}
}

func (s *TopicAlarmCollect) Range(f func(key string, entry *TopicAlarmEntry)) {
	tmp := map[string]*TopicAlarmEntry{}
	s.mu.RLock()
	for k, v := range s.m {
		tmp[k] = v
	}
	s.mu.RUnlock()

	for k, v := range tmp {
		f(k, v)
	}
}

func (s *TopicAlarmCollect) collect() {
	topics, err := dao.Topic{}.Query(&dao.Query{})
	if err != nil {
		log.Println("Query topic failed:", err)
		return
	}

	topicAlarmCache := map[string]*TopicAlarmEntry{}
	for _, topic := range topics {
		entry := &TopicAlarmEntry{
			Alarm:      topic.Alarm,
			AlarmRetry: topic.AlarmRetry,
		}
		topicAlarmCache[topic.GetQueueTopic()] = entry
	}

	s.mu.Lock()
	s.m = topicAlarmCache
	s.mu.Unlock()
}
