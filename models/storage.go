package models

import (
	"time"

	"github.com/huajiao-tv/dashboard/dao"
	"github.com/jinzhu/gorm"
	"github.com/youlu-cn/ginp"
)

type Storage struct {
	ginp.Model
	System         string        `json:"system" gorm:"unique_index:idx_system_host_port"`
	Type           string        `json:"type" gorm:"default:'redis'"`
	Host           string        `json:"host" gorm:"unique_index:idx_system_host_port"`
	Port           int           `json:"port" gorm:"unique_index:idx_system_host_port"`
	Password       string        `json:"password"`
	MaxConnNum     int           `json:"max_conn_num" gorm:"default:'100'"`
	MaxIdleNum     int           `json:"max_idle_num" gorm:"default:'100'"`
	MaxIdleSeconds time.Duration `json:"max_idle_seconds" gorm:"default:'3000000000'"`
	Status         int           `json:"status"`
	Describe       string        `json:"desc" gorm:"column:description"`

	// operator
	Comment  string `form:"comment"`
	Operator string `form:"author"`
}

func (m Storage) TableName() string {
	return "storage"
}

func (m Storage) Create() (err error) {
	tx := dao.DB.Begin()
	if err = tx.Error; err != nil {
		return
	}
	if err = tx.Create(&m).Error; err != nil {
		goto Rollback
	}
	if err = tx.Model(&System{}).Where("name = ?", m.System).
		Update("storage_count", gorm.Expr("storage_count + ?", 1)).Error; err != nil {
		goto Rollback
	}
	return tx.Commit().Error

Rollback:
	if tx.Rollback().Error != nil {
		// todo: add log
	}
	return
}

func (m Storage) Update() error {
	return dao.DB.Save(&m).Error
}

func (m Storage) Get(id uint64) (v *Storage, err error) {
	db := dao.DB.Model(&m).Where("id = ?", id).First(&m)
	return &m, db.Error
}

func (m Storage) Query(query *Query) (v []*Storage, err error) {
	db := dao.DB.Model(&m)
	if query.System != "" {
		db = db.Where("system = ?", query.System)
	}
	db = db.Find(&v)
	return v, db.Error
}

func (m *Storage) FindOne(query map[string]string) error {
	db := dao.DB.Model(&m)
	for _, k := range []string{"system", "host", "port"} {
		if v, ok := query[k]; ok {
			db = db.Where(k+" = ?", v)
		}
	}
	db = db.First(&m)
	return db.Error
}

func (m Storage) FindAll() (v []*Storage, err error) {
	db := dao.DB.Model(&m)
	db = db.Where("status = ?", 0)

	db = db.Find(&v)
	return v, db.Error
}
