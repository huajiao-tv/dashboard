package dao

import (
	"github.com/huajiao-tv/dashboard/config"
	"github.com/youlu-cn/ginp"
)

type Queue struct {
	ginp.Model
	Name       string `binding:"required" json:"name" gorm:"unique_index"`
	Describe   string `binding:"required" json:"desc" gorm:"column:description"`
	Password   string `binding:"required" json:"password"`
	TopicCount int    `json:"topics"`

	// operator
	Comment  string `binding:"required" json:"comment"`
	Operator string `json:"author"`
}

func (m Queue) TableName() string {
	return "queue"
}

func (m Queue) Create() error {
	return config.MySQL.Create(&m).Error
}

func (m Queue) Query(query *Query) (v []*Queue, err error) {
	db := config.MySQL.Model(&m).Find(&v)
	return v, db.Error
}

func (m Queue) Delete() (err error) {
	return config.MySQL.Delete(&m).Error
}

func (m Queue) Get(id uint64) (v *Queue, err error) {
	db := config.MySQL.Model(&m).Where("id = ?", id).First(&m)
	return &m, db.Error
}

func (m Queue) GetByName(name string) (v *Queue, err error) {
	db := config.MySQL.Model(&m).Where("name = ?", name).First(&m)
	return &m, db.Error
}

func (m Queue) Update() error {
	return config.MySQL.Save(&m).Error
}

func (m Queue) FindAllByNames(names []string) ([]*Queue, error) {
	var v []*Queue
	db := config.MySQL.Model(&m).Where("name in (?)", names).Find(&v)
	return v, db.Error
}

func (m Queue) FindBlankQueue() ([]*Queue, error) {
	var v []*Queue
	db := config.MySQL.Model(&m).Where("topic_count = ?", 0).Find(&v)
	return v, db.Error
}

func (m Queue) FindAllByIds(ids []int) ([]*Queue, error) {
	var v []*Queue
	db := config.MySQL.Model(&m).Where("id in (?)", ids).Find(&v)
	if db.Error != nil {
		return nil, db.Error
	}
	return v, nil
}
