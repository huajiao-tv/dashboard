package dao

import (
	"fmt"
	"strings"
	"time"

	"github.com/huajiao-tv/dashboard/config"
	"github.com/youlu-cn/ginp"
)

type TopicHistory struct {
	ginp.Model
	Queue         string `json:"queue" gorm:"unique_index:uix_queue_topic"`
	Topic         string `json:"topic" gorm:"unique_index:uix_queue_topic"`
	Length        int64  `json:"length"`
	RetryLength   int64  `json:"retry_length"`
	TimeoutLength int64  `json:"timeout_length"`

	//
	DataType int `binding:"required" json:"type" form:"type" gorm:"-"`
}

func (m TopicHistory) TableName() string {
	switch m.DataType {
	case DayData:
		return fmt.Sprintf("topic_day_history_%s", time.Now().Local().Format("200601"))
	case MonthData:
		return fmt.Sprintf("topic_month_history_%s", time.Now().Local().Format("200601"))
	default:
		return fmt.Sprintf("topic_history_%s", time.Now().Local().Format("200601"))
	}
}

func (m TopicHistory) GetQueueTopic() string {
	return fmt.Sprintf("%v/%v", m.Queue, m.Topic)
}

func (m TopicHistory) CreateBatch(data map[string]*TopicHistory) error {
	if len(data) == 0 {
		return nil
	}
	var vals []interface{}
	var valStrings []string
	for _, stat := range data {
		vals = append(vals, stat.Queue, stat.Topic, stat.Length, stat.RetryLength)
		valStrings = append(valStrings, "(?, ?, ?, ?)")
	}
	sqlCmd := fmt.Sprintf("INSERT INTO %s (queue, topic, length, retry_length) VALUES %s",
		m.TableName(), strings.Join(valStrings, ","))
	return config.MySQL.Exec(sqlCmd, vals...).Error
}
