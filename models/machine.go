package models

import (
	"strings"

	"github.com/huajiao-tv/dashboard/dao"
	"github.com/jinzhu/gorm"
	"github.com/youlu-cn/ginp"
)

type Machine struct {
	ginp.Model
	System string `binding:"required" json:"system" gorm:"unique_index:idx_ip_system"`
	IP     string `json:"ip" gorm:"unique_index:idx_ip_system"`
	Status int    `json:"status"`

	// operator
	Comment  string `binding:"required" json:"comment"`
	Operator string `json:"author"`

	// only for request form
	Cron bool   `json:"cron" gorm:"-"`
	IPs  string `binding:"required" json:"ips" gorm:"-"`
}

func (m Machine) TableName() string {
	if m.Cron {
		return "machine_cron"
	}
	return "machine"
}

func (m Machine) Create() (err error) {
	ips := strings.Split(m.IPs, "\n")
	machines := make([]*Machine, 0, len(ips))
	for _, ip := range ips {
		machines = append(machines, &Machine{
			System:   m.System,
			IP:       ip,
			Comment:  m.Comment,
			Operator: m.Operator,
			Cron:     m.Cron,
		})
	}

	col := "machine_count"
	if m.Cron {
		col = "cron_machine_count"
	}
	tx := dao.DB.Begin()
	if err = tx.Error; err != nil {
		return
	}
	for _, machine := range machines {
		if err = tx.Table(machine.TableName()).Create(machine).Error; err != nil {
			goto Rollback
		}
	}
	if err = tx.Model(&System{}).Where("name = ?", m.System).
		Update(col, gorm.Expr(col+" + ?", len(machines))).Error; err != nil {
		goto Rollback
	}
	return tx.Commit().Error

Rollback:
	if tx.Rollback().Error != nil {
		// todo: add log
	}
	return
}

func (m Machine) Delete() (err error) {
	col := "machine_count"
	if m.Cron {
		col = "cron_machine_count"
	}
	tx := dao.DB.Begin()
	if err = tx.Error; err != nil {
		return
	}
	if err = tx.Table(m.TableName()).Delete(&m).Error; err != nil {
		goto Rollback
	}
	if err = tx.Model(&System{}).Where("name = ?", m.System).
		Update(col, gorm.Expr(col+" - 1")).Error; err != nil {
		goto Rollback
	}
	return tx.Commit().Error

Rollback:
	if tx.Rollback().Error != nil {
		// todo: add log
	}
	return
}

func (m Machine) Get(id uint64) (v *Machine, err error) {
	db := dao.DB.Model(&m).Where("id = ?", id).First(&m)
	return &m, db.Error
}

func (m Machine) Query(query *Query) (v []*Machine, err error) {
	db := dao.DB.Table(m.TableName())
	if query.Keyword != "" {
		db = db.Where("ip in (?)", strings.Split(query.Keyword, "\n"))
	}
	if query.System != "" {
		db = db.Where("system = ?", query.System)
	}
	db = db.Find(&v)
	return v, db.Error
}
