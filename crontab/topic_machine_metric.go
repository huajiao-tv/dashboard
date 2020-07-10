package crontab

import (
	"fmt"
	"log"
	"math"
	"sync"

	"github.com/huajiao-tv/dashboard/keeper"
)

type TopicMachineMetricsCollect struct {
	mu      sync.RWMutex
	metrics map[string]*TopicMachineMetric
}

func newTopicMachineMetricsCollect() *TopicMachineMetricsCollect {
	return &TopicMachineMetricsCollect{
		metrics: make(map[string]*TopicMachineMetric),
	}
}

type TopicMachineMetric struct {
	Node        string
	ConsumeLe50 float64
	ConsumeLe90 float64
	ConsumeLe99 float64
	Succ        float64
	Fail        float64
	SuccQpm     float64
	FailQpm     float64
	Question    string
}

func (s *TopicMachineMetricsCollect) GetMetrics() map[string]*TopicMachineMetric {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.metrics
}

func (s *TopicMachineMetricsCollect) Get(queue, topic, node string) *TopicMachineMetric {
	key := fmt.Sprintf("%v/%v-%v", queue, topic, node)
	s.mu.RLock()
	data, ok := s.metrics[key]
	s.mu.RUnlock()
	if ok {
		return data
	}
	return &TopicMachineMetric{Node: node}
}

func (s *TopicMachineMetricsCollect) getMetric() map[string]*TopicMachineMetric {
	data := make(map[string]*TopicMachineMetric)
	nodes, err := keeper.GetNodeList()
	if err != nil {
		log.Printf("CollectQueueStats getNodeList failed: %v", err)
		return data
	}

	metricsBatch, _ := getMetricsBatch(nodes)
	for node, metricsEach := range metricsBatch {
		if metricsEach["pepperbus_job_consume_milliseconds"] == nil &&
			metricsEach["pepperbus_job_fail_count"] == nil &&
			metricsEach["pepperbus_job_success_count"] == nil {
			continue
		}
		for _, metricEach := range metricsEach["pepperbus_job_consume_milliseconds"].GetMetric() {
			var queue, topic string
			for _, l := range metricEach.Label {
				if l.GetName() == "queue" {
					queue = l.GetValue()
				} else if l.GetName() == "topic" {
					topic = l.GetValue()
				}
			}
			key := fmt.Sprintf("%v/%v-%v", queue, topic, node)
			tme := data[key]
			if tme == nil {
				data[key] = &TopicMachineMetric{Node: node}
				tme = data[key]
			}
			for _, le := range metricEach.GetSummary().GetQuantile() {
				if le.GetQuantile()-0.5 < 0.0001 && !math.IsNaN(le.GetValue()) {
					tme.ConsumeLe50 = le.GetValue()
				} else if le.GetQuantile()-0.9 < 0.001 && !math.IsNaN(le.GetValue()) {
					tme.ConsumeLe90 = le.GetValue()
				} else if le.GetQuantile()-0.99 < 0.001 && !math.IsNaN(le.GetValue()) {
					tme.ConsumeLe99 = le.GetValue()
				}
			}
		}
		for _, metricEach := range metricsEach["pepperbus_job_success_count"].GetMetric() {
			var queue, topic string
			for _, l := range metricEach.Label {
				if l.GetName() == "queue" {
					queue = l.GetValue()
				} else if l.GetName() == "topic" {
					topic = l.GetValue()
				}
			}
			key := fmt.Sprintf("%v/%v-%v", queue, topic, node)
			tme := data[key]
			if tme == nil {
				tme = &TopicMachineMetric{Node: node}
				data[key] = tme
			}
			tme.Succ = metricEach.GetCounter().GetValue()
		}
		for _, metricEach := range metricsEach["pepperbus_job_fail_count"].GetMetric() {
			var queue, topic string
			for _, l := range metricEach.Label {
				if l.GetName() == "queue" {
					queue = l.GetValue()
				} else if l.GetName() == "topic" {
					topic = l.GetValue()
				}
			}
			key := fmt.Sprintf("%v/%v-%v", queue, topic, node)
			tme := data[key]
			if tme == nil {
				data[key] = &TopicMachineMetric{Node: node}
				tme = data[key]
			}
			tme.Fail = metricEach.GetCounter().GetValue()
		}
	}
	return data
}

func (s *TopicMachineMetricsCollect) collect() {
	data := s.getMetric()
	old := s.getMetric()
	for k, v := range data {
		if oldV := old[k]; oldV != nil {
			v.SuccQpm = v.Succ - oldV.Succ
			v.FailQpm = v.Fail - oldV.Fail
		}
	}
	s.mu.Lock()
	s.metrics = data
	s.mu.Unlock()
}
