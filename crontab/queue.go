package crontab

import (
	"log"
	"sync"

	"github.com/huajiao-tv/dashboard/dao"
	"github.com/huajiao-tv/dashboard/keeper"
)

type QueueCollect struct {
	mu sync.RWMutex
	m  map[string]*dao.QueueHistory
}

func newQueueCollect() *QueueCollect {
	return &QueueCollect{
		m: map[string]*dao.QueueHistory{},
	}
}

func (s *QueueCollect) Get(queue string) *dao.QueueHistory {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if data, ok := s.m[queue]; ok {
		return data
	}
	return &dao.QueueHistory{}
}

func (s *QueueCollect) collect() {
	nodes, err := keeper.GetNodeList()
	if err != nil {
		log.Printf("CollectQueueStats getNodeList failed: %v", err)
		return
	}

	nodeChan := make(chan *GatewayStatsResp, len(nodes))
	var wg sync.WaitGroup
	for _, node := range nodes {
		wg.Add(1)
		go func(node string) {
			defer wg.Done()
			stats, err := getGatewayStats(node)
			if err != nil {
				log.Printf("GetGatewayStats failed: %v", err)
				return
			}
			nodeChan <- stats
		}(node)
	}
	wg.Wait()

	// merge
	if len(nodeChan) == 0 {
		return
	}
	results := make(map[string]*GatewayStatsItem)

Range:
	for {
		select {
		case stats := <-nodeChan:
			for queue, data := range stats.Data {
				if _, ok := results[queue]; !ok {
					results[queue] = data
				} else {
					results[queue].SuccessCount += data.SuccessCount
					results[queue].FailCount += data.FailCount
				}
			}
		default:
			break Range
		}
	}

	// calculate qps
	data := make(map[string]*dao.QueueHistory, len(results))
	for queue, stats := range results {
		qps := 0.0
		if sec := stats.TimePeriod.Seconds(); sec > 0 {
			qps = float64(stats.SuccessCount+stats.FailCount) / sec
		}
		data[queue] = &dao.QueueHistory{
			Queue:        queue,
			SuccessCount: stats.SuccessCount,
			FailCount:    stats.FailCount,
			AddQps:       qps,
		}
	}
	s.mu.Lock()
	s.m = data
	s.mu.Unlock()
	err = dao.QueueHistory{DataType: dao.HourData}.CreateBatch(data)
	if err != nil {
		log.Printf("QueueHistory CreateBatch failed: %v", err)
	}
}
