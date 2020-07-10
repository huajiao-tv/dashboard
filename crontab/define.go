package crontab

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/huajiao-tv/dashboard/config"
	"github.com/huajiao-tv/dashboard/dao"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"github.com/robfig/cron"
	"github.com/youlu-cn/ginp"
)

var (
	RedisState          = newRedisStateCollect()
	TaskState           = newTaskStatusCollect()
	QueueCollectStats   = newQueueCollect()
	TopicLengthStats    = newTopicLengthCollect()
	SystemJobExecStats  = newSystemJobExecCollect()
	TopicScriptsStatus  = newTopicScriptsStatusCheck()
	TopicMachineMetrics = newTopicMachineMetricsCollect()
	TopicAlarmStats     = newTopicAlarmCollect()
)

func Init() {
	c := cron.New()
	c.AddFunc("@every 1m", RedisState.Update)
	c.AddFunc("@every 10s", RedisState.collect)
	c.AddFunc("@every 1m", QueueCollectStats.collect)
	c.AddFunc("@every 1m", TopicLengthStats.collect)
	c.AddFunc("@every 1m", SystemJobExecStats.collect)
	c.AddFunc("@every 1m", TopicMachineMetrics.collect)
	c.AddFunc("@every 5m", TopicScriptsStatus.collect)
	c.AddFunc("@every 1m", TaskState.collect)
	c.AddFunc("@every 1m", TaskTestClear)
	c.AddFunc("@every 1m", TopicAlarmStats.collect)
	c.AddFunc("@daily", DBTableCron)

	c.Start()
}

type BaseResponse struct {
	ErrorCode int    `json:"error_code"`
	Error     string `json:"error"`
}

type MuxStatsResp struct {
	BaseResponse
	Node string
	Data map[string]map[string]struct {
		TimePeriod   time.Duration
		SuccessCount int64
		FailCount    int64
		MaxTime      time.Duration
		MinTime      time.Duration
		TotalTime    time.Duration
	} `json:"data"`
}

type GatewayStatsItem struct {
	TimePeriod   time.Duration
	SuccessCount int64
	FailCount    int64
}

type GatewayStatsResp struct {
	BaseResponse
	Data map[string]*GatewayStatsItem `json:"data"`
}

type TopicLengthResp struct {
	BaseResponse
	Data struct {
		Normal  int64
		Retry   int64
		Timeout int64
	} `json:"data"`
}

type RetryErrorResp struct {
	BaseResponse
	Data []*struct {
		Time     string
		Jobs     string
		SysCode  string
		SysError string
		UserCode string
		UserMsg  string
	} `json:"data"`
}

type MachineStatsResp struct {
	BaseResponse
	Data struct {
		CPU    float64
		Load   float64
		Memory float64
	} `json:"data"`
}

type SendToCgiReq struct {
	Queue string `binding:"required" json:"queue"`
	Topic string `binding:"required" json:"topic"`
	Node  string `binding:"required" json:"node"`
	Type  string `binding:"required" json:"type"`
	Value string `binding:"required" json:"value"`
}

type SendToTopicResp struct {
	BaseResponse
	Data interface{} `json:"data"`
}

func SendToCgi(req *SendToCgiReq) (interface{}, error) {
	uri := fmt.Sprintf("http://%v/SendToCgi", req.Node)
	form := url.Values{
		"queue": []string{req.Queue},
		"topic": []string{req.Topic},
		"type":  []string{req.Type},
		"value": []string{req.Value},
	}
	resp, err := config.HttpClient.PostForm(uri, form)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	val := SendToTopicResp{}
	err = json.Unmarshal(body, &val)
	if err != nil {
		return nil, err
	}
	return val.Data, nil
}

func getMuxStats(node string) (*MuxStatsResp, error) {
	uri := fmt.Sprintf("http://%v/GetMuxStats", node)
	resp, err := config.HttpClient.Get(uri)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	muxStats := MuxStatsResp{}
	err = json.Unmarshal(body, &muxStats)
	if err != nil {
		return nil, err
	}
	return &muxStats, nil
}

func getGatewayStats(node string) (*GatewayStatsResp, error) {
	u := fmt.Sprintf("http://%v/GetGatewayStats", node)
	resp, err := config.HttpClient.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	gatewayStats := GatewayStatsResp{}
	err = json.Unmarshal(body, &gatewayStats)
	if err != nil {
		return nil, err
	}
	return &gatewayStats, nil
}

func GetTopicLength(node, queue, topic string) (*TopicLengthResp, error) {
	u := fmt.Sprintf("http://%v/GetQueueLength?queue=%v&topic=%v", node, queue, topic)
	resp, err := config.HttpClient.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	topicLength := TopicLengthResp{}
	err = json.Unmarshal(body, &topicLength)
	if err != nil {
		return nil, err
	}
	return &topicLength, nil
}

func GetCGIError(node, queue, topic string, count int) (*RetryErrorResp, error) {
	u := fmt.Sprintf("http://%v/GetCGIError?queue=%v&topic=%v&count=%d", node, queue, topic, count)
	resp, err := config.HttpClient.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	retryErrors := RetryErrorResp{}

	err = json.Unmarshal(body, &retryErrors)
	if err != nil {
		return nil, err
	}
	return &retryErrors, nil
}

type TopicOperateResp struct {
	BaseResponse
	Data int64 `json:"data"`
}

func CleanTopicRetry(node, queue, topic string) (int64, error) {
	u := fmt.Sprintf("http://%v/CleanFailedJobs", node)
	form := url.Values{
		"queue": []string{queue},
		"topic": []string{topic},
	}
	resp, err := config.HttpClient.PostForm(u, form)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	ret := TopicOperateResp{}
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return 0, err
	}
	return ret.Data, nil
}

func CleanStatistics(node, queue, topic string) (int64, error) {
	u := fmt.Sprintf("http://%v/CleanStatistics", node)
	form := url.Values{
		"queue": []string{queue},
		"topic": []string{topic},
	}
	resp, err := config.HttpClient.PostForm(u, form)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	ret := TopicOperateResp{}
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return 0, err
	}
	return ret.Data, nil
}

func RecoverTopic(node, queue, topic string, count int) (int64, error) {
	u := fmt.Sprintf("http://%v/Recover", node)
	form := url.Values{
		"queue": []string{queue},
		"topic": []string{topic},
		"count": []string{strconv.Itoa(count)},
	}
	resp, err := config.HttpClient.PostForm(u, form)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	ret := TopicOperateResp{}
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return 0, err
	}
	return ret.Data, nil
}

func getMachineStats(node string) (*MachineStatsResp, error) {
	u := fmt.Sprintf("http://%v/GetMachineStats", node)
	resp, err := config.HttpClient.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	machineStats := MachineStatsResp{}
	err = json.Unmarshal(body, &machineStats)
	if err != nil {
		return nil, err
	}
	return &machineStats, nil
}

type TaskDebugData struct {
	Time string
	Jobs string
}

//获取cgi错误的 job
// demo http://localhost:19840/GetCGITaskDebug?queue=queue1&count=10
func GetCGITaskDebug(node, queue string, count int) ([]TaskDebugData, error) {
	u := fmt.Sprintf("http://%v/GetCGITaskDebug?queue=%s&count=%d", node, queue, count)
	resp, err := config.HttpClient.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	machineStats := struct {
		BaseResponse
		Data []TaskDebugData `json:"data"`
	}{}
	err = json.Unmarshal(body, &machineStats)
	if err != nil {
		return nil, err
	}
	return machineStats.Data, nil
}

func getMetricsBatch(nodes []string) (map[string]map[string]*dto.MetricFamily, error) {
	var g sync.WaitGroup
	metrics := make(map[string]map[string]*dto.MetricFamily)
	var mu sync.Mutex
	for _, node := range nodes {
		g.Add(1)
		go func(n string) {
			defer g.Done()
			m, err := getMetrics(n)
			if err != nil {
				return
			}
			mu.Lock()
			metrics[n] = m
			mu.Unlock()
		}(node)
	}
	g.Wait()
	return metrics, nil
}

func getMetrics(node string) (map[string]*dto.MetricFamily, error) {
	u := fmt.Sprintf("http://%v/metrics", node)
	resp, err := config.HttpClient.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var parser expfmt.TextParser
	res, err := parser.TextToMetricFamilies(strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	return res, nil
}

func DBTableCron() {
	queueHistory := dao.QueueHistory{
		Model: ginp.Model{
			CreatedAt: time.Now().Add(time.Hour * 24).Local(),
		},
		DataType: dao.HourData,
	}
	topicHistory := dao.TopicHistory{
		Model: ginp.Model{
			CreatedAt: time.Now().Add(time.Hour * 24).Local(),
		},
		DataType: dao.HourData,
	}
	if !config.MySQL.HasTable(&queueHistory) {
		config.MySQL.CreateTable(&queueHistory)
	}
	if !config.MySQL.HasTable(&topicHistory) {
		config.MySQL.CreateTable(&topicHistory)
	}
}
