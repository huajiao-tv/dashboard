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
	if !config.Postgres.HasTable(&User{}) {
		config.Postgres.CreateTable(&User{})
	}
	if !config.Postgres.HasTable(&System{}) {
		config.Postgres.CreateTable(&System{})
	}

	if !config.Postgres.HasTable(&Queue{}) {
		config.Postgres.CreateTable(&Queue{})
	}
	if !config.Postgres.HasTable(&QueueHistory{DataType: HourData}) {
		config.Postgres.CreateTable(&QueueHistory{DataType: HourData})
	}
	//config.Postgres.CreateTable(&QueueHistory{DataType: DayData})
	//config.Postgres.CreateTable(&QueueHistory{DataType: MonthData})

	if !config.Postgres.HasTable(&Machine{}) {
		config.Postgres.CreateTable(&Machine{})
	}
	if !config.Postgres.HasTable(&Machine{Cron: true}) {
		config.Postgres.CreateTable(&Machine{Cron: true})
	}

	if !config.Postgres.HasTable(&Topic{}) {
		config.Postgres.CreateTable(&Topic{})
	}
	if !config.Postgres.HasTable(&TopicHistory{DataType: HourData}) {
		config.Postgres.CreateTable(&TopicHistory{DataType: HourData})
	}

	if !config.Postgres.HasTable(&Storage{}) {
		config.Postgres.CreateTable(&Storage{})
	}
	if !config.Postgres.HasTable(&Task{}) {
		config.Postgres.CreateTable(&Task{})
	}
	if !config.Postgres.HasTable(&TaskTest{}) {
		config.Postgres.CreateTable(&TaskTest{})
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
