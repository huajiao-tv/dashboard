package crontab

import (
	"fmt"
	"log"
	"math/rand"
	"sync"

	"github.com/huajiao-tv/dashboard/dao"
	"github.com/huajiao-tv/dashboard/keeper"
)

type TopicLengthCollect struct {
	mu sync.RWMutex
	m  map[string]*dao.TopicHistory
}

func newTopicLengthCollect() *TopicLengthCollect {
	return &TopicLengthCollect{
		m: map[string]*dao.TopicHistory{},
	}
}

func (s *TopicLengthCollect) GetByKey(key string) *dao.TopicHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if data, ok := s.m[key]; ok {
		return data
	}
	return &dao.TopicHistory{}
}

func (s *TopicLengthCollect) Get(queue, topic string) *dao.TopicHistory {
	key := fmt.Sprintf("%v/%v", queue, topic)
	return s.GetByKey(key)
}

func (s *TopicLengthCollect) collect() {
	nodes, err := keeper.GetNodeList()
	if err != nil {
		log.Println("CollectTopicLength getNodeList failed:", err)
		return
	}
	if len(nodes) == 0 {
		log.Println("CollectTopicLength getNodeList get empty node,skip!")
		return
	}
	topics, err := dao.Topic{}.Query(&dao.Query{})
	if err != nil {
		log.Println("Query topic failed:", err)
		return
	}

	topicChan := make(chan *dao.TopicHistory, len(topics))
	for _, topic := range topics {
		node := nodes[rand.Intn(len(nodes))]
		info, err := GetTopicLength(node, topic.Queue, topic.Name)
		if err != nil {
			log.Println("getTopicLength failed:", err)
			return
		}
		topicChan <- &dao.TopicHistory{
			Queue:         topic.Queue,
			Topic:         topic.Name,
			Length:        info.Data.Normal,
			RetryLength:   info.Data.Retry,
			TimeoutLength: info.Data.Timeout,
		}
	}

	// save
	if len(topicChan) == 0 {
		return
	}
	data := make(map[string]*dao.TopicHistory, len(topicChan))

Range:
	for {
		select {
		case stats := <-topicChan:
			key := fmt.Sprintf("%v/%v", stats.Queue, stats.Topic)
			data[key] = stats
		default:
			break Range
		}
	}

	s.mu.Lock()
	s.m = data
	s.mu.Unlock()
	err = dao.TopicHistory{DataType: dao.HourData}.CreateBatch(data)
	if err != nil {
		log.Println("TopicHistory CreateBatch failed:", err)
	}
}
