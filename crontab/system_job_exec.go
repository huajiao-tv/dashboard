package crontab

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/huajiao-tv/dashboard/dao"
	"github.com/huajiao-tv/dashboard/models"
)

type SystemJobExecCollect struct {
	sync.RWMutex

	success map[string]uint64
	fail    map[string]uint64
}

func (s *SystemJobExecCollect) GetCount(system string) (uint64, uint64) {
	s.RLock()
	defer s.RUnlock()
	return s.success[system], s.fail[system]
}

func (s *SystemJobExecCollect) collect() {
	m := models.Task{}
	tasks, err := m.Query()
	if err != nil {
		return
	}
	success := make(map[string]uint64)
	fail := make(map[string]uint64)
	for _, task := range tasks {
		key := fmt.Sprintf("job:times:%s:%d", time.Now().Format("20060102"), task.ID)
		m, err := dao.RC.HGetAll(key).Result()
		if err != nil {
			continue
		}
		sc, _ := strconv.ParseUint(m["success"], 10, 0)
		fc, _ := strconv.ParseUint(m["failed"], 10, 0)
		success[task.System] += sc
		fail[task.System] += fc
	}
	s.Lock()
	s.success = success
	s.fail = fail
	s.Unlock()
}
