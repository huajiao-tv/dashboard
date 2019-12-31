package dao

import (
	"time"

	"github.com/huajiao-tv/dashboard/config"
)

type JobResult struct {
	Id           uint   `gorm:"primary_key"`
	JobName      string `json:"job_name"`
	DispatchID   string `json:"dispatch_id"`
	JobCompleted bool
	GroupID      float64   `json:"group_id"`
	AgentNode    string    `json:"agent_node"`
	StartedAt    time.Time `json:"started_at"`
	FinishedAt   time.Time `json:"finished_at"`
	Status       int       `json:"status"`
	Output       string    `json:"output,omitempty" gorm:"column:output_data"`
}

type jobLog struct {
	Id         uint      `json:"id"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
	Output     string    `json:"output" gorm:"column:output_data"`
}

func (jobLog) TableName() string {
	return "job_result"
}

func (JobResult) TableName() string {
	return "job_result"
}

func (m JobResult) Get(name string) []*JobResult {
	var v []*JobResult
	config.MySQL.Model(&m).Where("job_name = ?", name).Order("id desc").Find(&v)
	return v
}

func (m JobResult) Query(taskID uint64, node string, page int64, sort string, limit int64, search string) *LogResult {
	var logs []*jobLog
	var count int

	ti := time.Now().AddDate(0, 0, -7)
	db := config.MySQL.Model(&m)
	if search == "" {
		db = db.Select("id,started_at,finished_at,output_data").
			Where("started_at >= ? and agent_node=? and job_name=?", ti, node, taskID)
	} else {
		db = db.Select("id,started_at,finished_at,output_data").
			Where("started_at >= ? and agent_node=? and job_name=? and output_data like ?", ti, node, taskID, "%"+search+"%")
	}
	if page == 1 {
		db = db.Count(&count)
	}
	db.Order("id " + sort).Offset((page - 1) * limit).Limit(limit).Find(&logs)
	return &LogResult{
		Log:   logs,
		Count: count,
	}
}

func (m JobResult) Count(taskID uint64, node, search string) (total int64) {
	begin := time.Now().AddDate(0, 0, -7)
	db := config.MySQL.Model(&m)
	if search == "" {
		db.Where("started_at >= ? and agent_node=? and job_name=?", begin, node, taskID).Count(&total)
	} else {
		db.Where("started_at >= ? and agent_node=? and job_name=? and output_data like '%?%'", begin, node, taskID, search).Count(&total)
	}
	return total
}
