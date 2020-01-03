package models

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/huajiao-tv/dashboard/config"
	"github.com/huajiao-tv/dashboard/dao"
	"github.com/huajiao-tv/peppercron/logic"
)

const (
	DefaultFrontPort = ":12306"
	DefaultAdminPort = ":12307"
	leaderStr        = "http://%v" + DefaultAdminPort + "/Leader"
	addJobStr        = "http://%v" + DefaultFrontPort + "/v1/jobs"
	deleteJobStr     = "http://%v" + DefaultFrontPort + "/v1/jobs/%v"
)

type Task struct {
}

func NewTask() *Task {
	return &Task{}
}

func (m Task) Add(task *dao.Task) error {
	var (
		err                  error
		runType              dao.RunType
		envs                 []map[string]interface{}
		environmentVariables []string
		executorParams       []string
	)
	var (
		rt    = logic.JobTypeSchedule
		delay = time.Duration(0)
		times = int64(0)
	)

	if err = json.Unmarshal([]byte(task.RunType), &runType); err != nil {
		return err
	}
	if runType.Type == 2 {
		rt = logic.JobTypeTimes
		times = -1
		if delay, err = time.ParseDuration(runType.Value + runType.Unit); err != nil {
			return err
		}
	}
	switch logic.ExecutorType(task.Executor) {
	case logic.ExecutorTypeShell:
		executorParams = make([]string, 3)
		executorParams[0] = "sh"
		executorParams[1] = "-c"
		executorParams[2] = task.Exec
	case logic.ExecutorTypeGRPC:
		executorParams = make([]string, 1)
		executorParams[0] = task.GRPCHost
	}

	if err = json.Unmarshal([]byte(task.EnvVar), &envs); err != nil {
		return err
	}
	for _, val := range envs {
		environmentVariables = append(environmentVariables, fmt.Sprintf("%v=%v", val["key"], val["value"]))
	}

	jobConfig := logic.JobConfig{
		AgentJobConfig: logic.AgentJobConfig{
			Name:                 task.GetEtcdJobName(),
			Type:                 rt,
			Times:                times,
			TimesDelay:           delay,
			ExecutorTimeout:      time.Duration(task.TimeOut) * time.Second,
			ExecutorType:         logic.ExecutorType(task.Executor),
			ExecutorParams:       executorParams,
			EnvironmentVariables: environmentVariables,
			Schedule:             runType.String(),
			Extras: map[string]string{
				"Name": task.Name,
			},
		},
		Concurrency:   task.Concurrency,
		Tag:           task.System,
		DependentJobs: task.GetDepends(),
		Retries:       3,
	}
	if err = m.AddJob(m.GetLeaderIp(task), &jobConfig); err != nil {
		return err
	}
	return nil
}

func (m Task) AddTest(task *dao.Task, username string) error {
	var (
		runType              dao.RunType
		executorParams       []string
		envs                 []map[string]interface{}
		environmentVariables []string
	)

	taskDetail, _ := json.Marshal(task)
	taskTest := &dao.TaskTest{
		TaskId:     task.ID,
		TaskJobId:  fmt.Sprintf("%v-%v", task.ID, time.Now().Unix()),
		TaskDetail: string(taskDetail),
		OpUser:     username,
	}
	if err := taskTest.Create(); err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(task.RunType), &runType); err != nil {
		return err
	}
	switch logic.ExecutorType(task.Executor) {
	case logic.ExecutorTypeShell:
		executorParams = make([]string, 3)
		executorParams[0] = "sh"
		executorParams[1] = "-c"
		executorParams[2] = task.Exec
	case logic.ExecutorTypeGRPC:
		executorParams = make([]string, 1)
		executorParams[0] = task.GRPCHost
	}

	if err := json.Unmarshal([]byte(task.EnvVar), &envs); err != nil {
		return err
	}
	for _, val := range envs {
		environmentVariables = append(environmentVariables, fmt.Sprintf("%v=%v", val["key"], val["value"]))
	}

	jobConfig := logic.JobConfig{
		AgentJobConfig: logic.AgentJobConfig{
			Name:                 taskTest.TaskJobId,
			Type:                 logic.JobTypeTimes,
			ExecutorTimeout:      time.Duration(task.TimeOut) * time.Second,
			ExecutorType:         logic.ExecutorType(task.Executor),
			ExecutorParams:       executorParams,
			EnvironmentVariables: environmentVariables,
			Times:                1,
			Extras: map[string]string{
				"Name": task.Name,
			},
		},
		Concurrency:   task.Concurrency,
		Tag:           task.System,
		DependentJobs: []string{},
		Retries:       3,
	}
	jbStr, _ := json.Marshal(jobConfig)
	_, err := config.ETCDClient.Put(context.TODO(), fmt.Sprintf(logic.JobConfiguration, taskTest.TaskJobId), string(jbStr))
	if err != nil {
		return err
	}
	return nil
}

func (m Task) GetInfo(task *dao.Task) map[string]string {
	key := fmt.Sprintf("job:times:%s", task.GetEtcdJobName())
	result, err := config.RedisClient.HGetAll(key).Result()
	if err != nil {
		return map[string]string{}
	}
	key = fmt.Sprintf("job:times:%v:%s", time.Now().Format("20060102"), task.GetEtcdJobName())
	dayRes, err := config.RedisClient.HGetAll(key).Result()
	if err != nil {
		return map[string]string{}
	}
	for k, v := range dayRes {
		result["day_"+k] = v
	}
	return result
}

func (m Task) GetMachineIp(task *dao.Task) []string {
	machines, err := dao.Machine{Cron: true}.Query(&dao.Query{System: task.System})
	if err != nil {
		return nil
	}
	if len(machines) == 0 {
		return nil
	}
	ips := make([]string, len(machines))
	for _, m := range machines {
		ips = append(ips, m.IP)
	}
	return ips
}

func (m Task) GetLeaderIp(task *dao.Task) string {
	machineIp := m.GetMachineIp(task)
	if machineIp == nil || len(machineIp) == 0 {
		log.Println("empty machines", task.System)
		return ""
	}
	for _, ip := range machineIp {
		url := fmt.Sprintf(leaderStr, ip)
		res, err := http.Get(url)
		if err != nil {
			continue
		}
		body, err := ioutil.ReadAll(res.Body)
		res.Body.Close()
		data, err := simplejson.NewJson(body)
		leaderIp, err := data.Get("data").String()
		if err != nil {
			continue
		}
		if leaderIp == "" {
			continue
		}
		return leaderIp
	}
	log.Println("error to get leader,machines: ", machineIp)
	return ""
}

func (m Task) AddJob(leaderIp string, job *logic.JobConfig) error {
	url := fmt.Sprintf(addJobStr, leaderIp)
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	res, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		log.Println("add job failed", err.Error())
		return err
	}
	code := res.StatusCode
	if code != 200 {
		return errors.New(fmt.Sprintf("error request code:%v", code))
	}
	return nil
}

func (m Task) DeleteJob(leaderIp string, jobName string) error {
	url := fmt.Sprintf(deleteJobStr, leaderIp, jobName)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	code := res.StatusCode
	if code != 200 {
		return errors.New(fmt.Sprintf("error request code:%v", code))
	}
	return nil
}

func (m Task) StopTask(task *dao.Task) error {
	return m.DeleteJob(m.GetLeaderIp(task), task.GetEtcdJobName())
}

func (m Task) GetDependency() map[string][]string {
	tasks := dao.Task{}.GetAllWork()
	res := make(map[string][]string)
	for _, val := range tasks {
		var (
			depends  []interface{}
			children []interface{}
		)
		_ = json.Unmarshal([]byte(val.DependOn), &depends)
		for _, tv := range depends {
			res[tv.(map[string]interface{})["label"].(string)] = append(res[tv.(map[string]interface{})["label"].(string)], val.Name)
		}
		_ = json.Unmarshal([]byte(val.Children), &children)
		for _, tv := range children {
			res[tv.(map[string]interface{})["label"].(string)] = append(res[tv.(map[string]interface{})["label"].(string)], val.Name)
		}
	}
	return res
}

func (m Task) GetNodeWorkInfo(task *dao.Task, node string) (ret map[string]string) {
	key := fmt.Sprintf("job:times:node:%s:%s", task.GetEtcdJobName(), node)
	result, err := config.RedisClient.HGetAll(key).Result()
	if err != nil {
		return
	}
	key = fmt.Sprintf("job:times:node:%s:%s:%s", time.Now().Format("20060102"), task.GetEtcdJobName(), node)
	dayRes, err := config.RedisClient.HGetAll(key).Result()
	if err != nil {
		return
	}
	for k, v := range dayRes {
		result["day_"+k] = v
	}
	return result
}

func (m Task) GetNode(task *dao.Task) (map[string]interface{}, error) {
	taskPath := fmt.Sprintf(logic.JobDispatchRecord, task.GetEtcdJobName())
	res, err := config.ETCDClient.Get(context.TODO(), taskPath)
	if err != nil {
		return nil, err
	}
	if len(res.Kvs) <= 0 {
		return map[string]interface{}{}, nil
	}
	val := res.Kvs[0]
	var ret map[string]interface{}
	err = json.Unmarshal(val.Value, &ret)
	return ret, err
}
