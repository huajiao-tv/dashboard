package controllers

import (
	"time"

	"github.com/huajiao-tv/dashboard/dao"
)

type PageData struct {
	Items interface{} `json:"items"`
	Total int         `json:"total"`
}

type QueueConfig struct {
	QueueName string                  `json:"queue_name"`
	Password  string                  `json:"password"`
	Topic     map[string]*TopicConfig `json:"topic"`
}

type TopicConfig struct {
	Name           string         `json:"name"`
	Password       string         `json:"password"`
	RetryTimes     int            `json:"retry_times"`
	MaxQueueLength int            `json:"max_queue_length"`
	RunType        int            `json:"run_type"`
	Storage        *StorageConfig `json:"storage"`
	NumOfWorkers   uint64         `json:"num_of_workers"`
	ScriptEntry    string         `json:"script_entry"`
	CgiConfigKey   string         `json:"cgi_config_key"`
	HttpConfig     string         `json:"http_config"`
	IsRetry        bool           `json:"is_retry"`
	RetryTime      int64          `json:"retry_time"`
}

type TimeRetry struct {
	IsRetry   bool  `json:"switch"`
	RetryTime int64 `json:"durTime"`
}
type StorageConfig struct {
	Type           string        `json:"type"`
	Address        string        `json:"address"`
	Auth           string        `json:"auth"`
	MaxConnNum     int           `json:"max_conn_num"`
	MaxIdleNum     int           `json:"max_idle_num"`
	MaxIdleSeconds time.Duration `json:"max_idle_seconds"`
}

type MachineConfig struct {
	SharedId uint64
	Tags     []string
}

type DelForm struct {
	Id uint64 `binding:"required" form:"id" json:"id"`
}

type CronNodeConfig struct {
	Tags string `json:"tags"`
}

type TopicSummary struct {
	Id          uint64      `json:"id"`
	Queue       string      `json:"queue"`
	Name        string      `json:"name"`
	Describe    string      `json:"desc"`
	File        string      `json:"file"`
	Workers     int         `json:"workers"`
	Storage     string      `json:"storage"`
	Length      int64       `json:"length"`
	RetryLen    int64       `json:"retry"`
	TimeoutLen  int64       `json:"timeout"`
	Machines    interface{} `json:"machines"`
	System      string      `json:"system"`
	Status      uint8       `json:"status"`
	RunType     int         `json:"run_type"`
	Alarm       int         `json:"alarm"`
	AlarmRetry  int         `json:"alarm_retry"`
	HttpConfig  string      `json:"http_config"`
	TopicConfig string      `json:"topic_config"`
}

type TaskD struct {
	*dao.Task
	Statics map[string]string
}

type TaskN struct {
	dao.TaskDetail
	Statics map[string]string
}
