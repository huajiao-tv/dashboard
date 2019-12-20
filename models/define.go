package models

import (
	"log"

	"github.com/huajiao-tv/dashboard/config"
	"github.com/huajiao-tv/dashboard/dao"
	"github.com/huajiao-tv/dashboard/ldap"
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

func Init() {
	if !dao.DB.HasTable(&User{}) {
		dao.DB.CreateTable(&User{})
	}
	if !dao.DB.HasTable(&System{}) {
		dao.DB.CreateTable(&System{})
	}

	if !dao.DB.HasTable(&Queue{}) {
		dao.DB.CreateTable(&Queue{})
	}

	if !dao.DB.HasTable(&Machine{}) {
		dao.DB.CreateTable(&Machine{})
	}
	if !dao.DB.HasTable(&Machine{Cron: true}) {
		dao.DB.CreateTable(&Machine{Cron: true})
	}

	if !dao.DB.HasTable(&Topic{}) {
		dao.DB.CreateTable(&Topic{})
	}

	if !dao.DB.HasTable(&Storage{}) {
		dao.DB.CreateTable(&Storage{})
	}
	if !dao.DB.HasTable(&Task{}) {
		dao.DB.CreateTable(&Task{})
	}
	if !dao.DB.HasTable(&TaskTest{}) {
		dao.DB.CreateTable(&TaskTest{}).ModifyColumn("task_detal", "text")
	}

	if len(*config.AdminPass) > 0 {
		u := User{
			Name:     "admin",
			Password: *config.AdminPass,
			Avatar:   ldap.DefaultAvatar,
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
