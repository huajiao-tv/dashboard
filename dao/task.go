package dao

import (
	"encoding/json"
	"strconv"

	"github.com/huajiao-tv/dashboard/config"
	"github.com/jinzhu/gorm"
	"github.com/youlu-cn/ginp"
)

const (
	Cron = iota
	Dur
)

const (
	TaskWorking = iota
	TaskStop
	TaskRemove
)

type Task struct {
	ginp.Model
	System      string `binding:"required" json:"system"`
	Name        string `binding:"required" json:"name" gorm:"unique_index:task_name"`
	Describe    string `binding:"required" json:"desc" gorm:"column:description"`
	TimeOut     int    `json:"timeout" gorm:"default:500"`
	Concurrency int    `json:"concurrency" gorm:"default:'1'"`
	Executor    int    `json:"executor" gorm:"default:0"`
	GRPCHost    string `json:"grpc_host" gorm:"column:grpc_host"`
	Exec        string `json:"command"`
	RunType     string `json:"scheduler"`
	Status      int    `json:"status" gorm:"default:0"`                        // 0表示开启
	NodeStatus  int    `json:"nodeStatus" gorm:"column:node_status;default:0"` // 0表示正常，1表示无可用节点
	DependOn    string `json:"dependents"`
	Children    string `json:"children"`
	Operator    string `json:"author"`
	EnvVar      string `json:"envs"`
}

func (t Task) TableName() string {
	return "task"
}

func (t *Task) GetEtcdJobName() string {
	return strconv.FormatUint(t.ID, 10)
}

func (t *Task) Create() error {
	tx := config.MySQL.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	err := tx.Create(t).Error
	if err != nil {
		return err
	} else {
		if err = tx.Model(&System{}).Where("name = ?", t.System).
			Update("job_count", gorm.Expr("job_count + 1")).Error; err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
		return tx.Error
	}
}

func (t *Task) Get(id int64) *Task {
	config.MySQL.Where("id=?", id).First(t)
	return t
}

func (t *Task) Query() (v []*Task, err error) {
	db := config.MySQL.Model(t).Find(&v)
	return v, db.Error
}

func (t *Task) Update() error {
	tx := config.MySQL.Begin()
	err := tx.Save(t).Error
	if err != nil {
		return err
	}
	if t.Status == TaskRemove {
		if err := tx.Model(&System{}).Where("name = ?", t.System).
			Update("job_count", gorm.Expr("job_count - 1")).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return tx.Error
}

func (t Task) GetAllWork() []*Task {
	var tasks []*Task
	config.MySQL.Where("status=? or status=?", TaskWorking, TaskStop).Find(&tasks)
	return tasks
}

func (t Task) GetDepsTree() *DepsTree {
	children := t.getChildren()
	deps := &DepsTree{
		Task: &t,
	}
	if children == nil {
		return deps
	} else {
		for _, val := range children {
			deps.Children = append(deps.Children, val.GetDepsTree())
		}
	}
	return deps
}

// TODO 需要优化
func (t *Task) GetDepends() []string {
	var ret []string

	var depends []map[string]interface{}
	_ = json.Unmarshal([]byte(t.DependOn), &depends)
	for _, val := range depends {
		task := new(Task)
		config.MySQL.Where("id=?", val["value"].(float64)).First(task)
		ret = append(ret, task.GetEtcdJobName())

	}
	var children []map[string]interface{}
	_ = json.Unmarshal([]byte(t.Children), &children)
	if len(children) == 0 {
		return []string{}
	}
	for _, val := range children {
		task := new(Task)
		config.MySQL.Where("id=?", val["value"].(float64)).First(task)
		ret = append(ret, task.GetEtcdJobName())
	}
	return ret
}

func (t *Task) getChildren() (ret []*Task) {
	cs := t.GetDepends()
	if cs == nil {
		return
	}
	for _, val := range cs {
		task := new(Task)
		config.MySQL.Where("name=?", val).First(task)
		if task.ID == 0 {
			continue
		}
		ret = append(ret, task)
	}
	return ret
}
