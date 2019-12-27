package dao

import (
	"github.com/huajiao-tv/dashboard/config"
	"github.com/youlu-cn/ginp"
)

type TaskTest struct {
	ginp.Model
	TaskId     uint64
	TaskJobId  string
	Status     int `gorm:"default:1"`
	Result     string
	TaskDetail string `gorm:"type:text"`
	OpUser     string
}

func (TaskTest) TableName() string {
	return "task_tests"
}

func (m TaskTest) Create() error {
	return config.Postgres.Create(&m).Error
}

func (m TaskTest) Get(id uint64) *TaskTest {
	config.Postgres.Model(&m).Where("task_id = ?", id).Order("id desc").First(&m)
	return &m
}
