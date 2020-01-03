package dao

import (
	"log"

	"github.com/huajiao-tv/dashboard/config"
)

const (
	HourData = iota
	DayData
	MonthData
)

type Query struct {
	Keyword  string   `form:"keyword" json:"keyword"`
	System   string   `form:"system" json:"system"`
	Topic    string   `form:"topic" json:"topic"`
	Queue    string   `form:"queue" json:"queue"`
	Count    int      `form:"count" json:"count"`
	Machines []string `form:"machines" json:"machines"`
}

type TransModelQueue struct {
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	Password string `json:"password"`
	Comment  string `json:"comment"`
	Topics   []TransModelTopic
}

type TransModelTopic struct {
	System   string            `json:"system"`
	Queue    string            `json:"queue"`
	Name     string            `json:"name"`
	Desc     string            `json:"desc"`
	Password string            `json:"password"`
	Consume  string            `json:"consume"`
	Comment  string            `json:"comment"`
	Status   string            `json:"status"`
	Storage  TransModelStorage `json:"storage"`
}

type TransModelStorage struct {
	SType string `json:"type"`
	Host  string `json:"host"`
	Port  string `json:"port"`
}

type LogResult struct {
	Log   []*jobLog `json:"log"`
	Count int       `json:"count"`
}

type DepsTree struct {
	*Task
	Children []*DepsTree
}

type TaskDetail struct {
	*Task
	Nodes map[string]interface{}
}

type Node struct {
	Name  string
	Tasks []*Task
}

type RunType struct {
	Value string
	Type  int
	Unit  string
}

func (r *RunType) String() string {
	switch r.Type {
	case Cron:
		return r.Value
	case Dur:
		str := "@every " + r.Value + r.Unit
		return str
	}
	return ""
}

func Init() {
	if !config.MySQL.HasTable(&User{}) {
		config.MySQL.CreateTable(&User{})
	}
	if !config.MySQL.HasTable(&System{}) {
		config.MySQL.CreateTable(&System{})
	}

	if !config.MySQL.HasTable(&Queue{}) {
		config.MySQL.CreateTable(&Queue{})
	}
	if !config.MySQL.HasTable(&QueueHistory{DataType: HourData}) {
		config.MySQL.CreateTable(&QueueHistory{DataType: HourData})
	}
	//config.MySQL.CreateTable(&QueueHistory{DataType: DayData})
	//config.MySQL.CreateTable(&QueueHistory{DataType: MonthData})

	if !config.MySQL.HasTable(&Machine{}) {
		config.MySQL.CreateTable(&Machine{})
	}
	if !config.MySQL.HasTable(&Machine{Cron: true}) {
		config.MySQL.CreateTable(&Machine{Cron: true})
	}

	if !config.MySQL.HasTable(&Topic{}) {
		config.MySQL.CreateTable(&Topic{})
	}
	if !config.MySQL.HasTable(&TopicHistory{DataType: HourData}) {
		config.MySQL.CreateTable(&TopicHistory{DataType: HourData})
	}

	if !config.MySQL.HasTable(&Storage{}) {
		config.MySQL.CreateTable(&Storage{})
	}
	if !config.MySQL.HasTable(&Task{}) {
		config.MySQL.CreateTable(&Task{})
	}
	if !config.MySQL.HasTable(&TaskTest{}) {
		config.MySQL.CreateTable(&TaskTest{})
	}
	if !config.MySQL.HasTable(&JobResult{}) {
		config.MySQL.CreateTable(&JobResult{})
	}

	if len(config.GlobalConfig.Dashboard.AdminPassword) > 0 {
		u := User{
			Name:     "admin",
			Password: config.GlobalConfig.Dashboard.AdminPassword,
			Avatar:   config.DefaultAvatar,
			Email:    "",
			Type:     0,
			Roles:    Administration,
		}
		if t, err := u.Get(); t == nil && err != nil {
			if err := u.Create(); err != nil {
				log.Println("create admin error, admin may exists: ", err.Error())
			}
		} else {
			log.Println("create admin error, admin exists!")
		}
	}
}
