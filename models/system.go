package models

import (
	"github.com/huajiao-tv/dashboard/dao"
	"github.com/youlu-cn/ginp"
)

type System struct {
	ginp.Model
	Name             string `binding:"required" form:"name" json:"name" gorm:"unique_index"`
	Describe         string `binding:"required" form:"desc" json:"desc" gorm:"column:description"`
	MachineCount     int    `json:"machines"`
	StorageCount     int    `json:"storages"`
	TopicCount       int    `json:"topics"`
	CronMachineCount int    `json:"cron_machines"`
	JobCount         int    `json:"jobs"`

	// operator
	Comment  string `form:"comment"`
	Operator string `form:"author"`

	// cron job count
	JobSuccessCount uint64 `json:"job_success_count" gorm:"-"`
	JobFailCount    uint64 `json:"job_fail_count" gorm:"-"`
}

func (m System) TableName() string {
	return "system"
}

func (m System) Create() error {
	return dao.DB.Create(&m).Error
}

func (m System) Query(_ *Query) (v []*System, err error) {
	db := dao.DB.Model(&m).Find(&v)
	return v, db.Error
}
