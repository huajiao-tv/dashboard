package crontab

import (
	"encoding/json"
	"time"

	"github.com/huajiao-tv/dashboard/config"
	"github.com/huajiao-tv/dashboard/dao"
	"github.com/huajiao-tv/dashboard/models"
)

func TaskTestClear() {
	var testTasks []*dao.TaskTest
	config.MySQL.Where("status<?", 3).Find(&testTasks)
	for _, val := range testTasks {
		if val.CreatedAt.Add(300 * time.Second).Before(time.Now()) {
			task := new(dao.Task)
			if err := json.Unmarshal([]byte(val.TaskDetail), task); err != nil {
				continue
			}
			_ = models.NewTask().StopTask(task)
			config.MySQL.Model(&val).Updates(dao.TaskTest{Status: 3})
		}
	}
}
