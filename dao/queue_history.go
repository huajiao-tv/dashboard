package dao

import (
	"fmt"
	"strings"
	"time"

	"github.com/huajiao-tv/dashboard/config"
	"github.com/youlu-cn/ginp"
)

type QueueHistory struct {
	ginp.Model
	Queue        string  `json:"queue" gorm:"index:idx_queue"`
	SuccessCount int64   `json:"success_count"`
	FailCount    int64   `json:"fail_count"`
	AddQps       float64 `json:"add_qps"`

	//
	DataType int `binding:"required" json:"type" form:"type" gorm:"-"`
}

func (m QueueHistory) TableName() string {
	dt := m.CreatedAt
	if dt.Unix() == 0 {
		dt = time.Now().Local()
	}
	switch m.DataType {
	case DayData:
		return fmt.Sprintf("queue_day_history_%s", dt.Format("200601"))
	case MonthData:
		return fmt.Sprintf("queue_month_history_%s", dt.Format("200601"))
	default:
		return fmt.Sprintf("queue_history_%s", dt.Format("200601"))
	}
}

func (m QueueHistory) CreateBatch(data map[string]*QueueHistory) error {
	if len(data) == 0 {
		return nil
	}
	var vals []interface{}
	var valStrings []string
	for _, stat := range data {
		vals = append(vals, stat.Queue, stat.SuccessCount, stat.FailCount, stat.AddQps)
		valStrings = append(valStrings, "(?, ?, ?, ?)")
	}
	sqlCmd := fmt.Sprintf("INSERT INTO %s (queue, success_count, fail_count, add_qps) VALUES %s",
		m.TableName(), strings.Join(valStrings, ","))
	return config.MySQL.Exec(sqlCmd, vals...).Error
}
