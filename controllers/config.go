package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/huajiao-tv/dashboard/dao"
	"github.com/youlu-cn/ginp"
)

var Config = new(ConfigController)

type ConfigController struct {
}

func (c ConfigController) Group() string {
	return "config"
}

func (c ConfigController) TokenRequired(string) bool {
	return true
}

func (c ConfigController) ListHandler(req *ginp.Request) *ginp.Response {
	query := dao.Query{}
	if err := req.Bind(&query); err != nil {
		return ginp.ErrorResponse(http.StatusBadRequest, err)
	}
	if query.Queue != "" {
		topicConfs, err := c.getTopicConfig(query.Queue)
		if err != nil {
			return ginp.ErrorResponse(http.StatusInternalServerError, err)
		}
		v, err := json.Marshal(topicConfs)
		if err != nil {
			return ginp.ErrorResponse(http.StatusInternalServerError, err)
		}
		return ginp.DataResponse(string(v))
	}
	queueConfs, err := c.getQueueConfig()
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	return ginp.DataResponse(queueConfs)
}

func (c ConfigController) GetTagsHandler(_ *ginp.Request) *ginp.Response {
	v, err := c.getTagsConfig()
	if err != nil {
		return ginp.ErrorResponse(http.StatusInternalServerError, err)
	}
	return ginp.DataResponse(v)
}

func (c ConfigController) getTopicConfig(queue string) (map[string]*TopicConfig, error) {
	confMap := make(map[string]*TopicConfig)
	topics, err := dao.Topic{}.Query(&dao.Query{Queue: queue})
	if err != nil {
		return nil, err
	} else if len(topics) == 0 {
		return confMap, nil
	}
	for _, topic := range topics {
		storage, err := dao.Storage{}.Get(topic.Storage)
		if err != nil {
			return nil, err
		}
		retry := new(TimeRetry)
		err = json.Unmarshal([]byte(topic.TopicConfig), retry)
		if err != nil {
			retry.IsRetry = false
			retry.RetryTime = 86400
		}
		confMap[topic.Name] = &TopicConfig{
			Name:           topic.Name,
			Password:       topic.Password,
			RetryTimes:     topic.RetryTimes,
			MaxQueueLength: topic.MaxQueueLength,
			RunType:        topic.RunType,
			NumOfWorkers:   uint64(topic.NumOfWorkers),
			ScriptEntry:    topic.ConsumeFile,
			Storage: &StorageConfig{
				Type:           storage.Type,
				Address:        fmt.Sprintf("%s:%d", storage.Host, storage.Port),
				Auth:           storage.Password,
				MaxConnNum:     storage.MaxConnNum,
				MaxIdleNum:     storage.MaxIdleNum,
				MaxIdleSeconds: storage.MaxIdleSeconds,
			},
			CgiConfigKey: topic.CgiConfig,
			HttpConfig:   topic.HttpConfig,
			IsRetry:      retry.IsRetry,
			RetryTime:    retry.RetryTime,
		}
	}
	return confMap, nil
}

func (c ConfigController) getAllQueueConfig() (string, error) {
	confMap := make(map[string]*QueueConfig)
	topics, err := dao.Topic{}.Query(&dao.Query{})
	if err != nil {
		return "", err
	}
	queues, err := dao.Queue{}.Query(&dao.Query{})
	if err != nil {
		return "", err
	}
	queuesMap := make(map[string]*dao.Queue)
	for _, queue := range queues {
		queuesMap[queue.Name] = queue
	}
	storages := make(map[uint64]*dao.Storage)
	for _, topic := range topics {
		queue, ok := queuesMap[topic.Queue]
		if !ok {
			continue
		}
		qc := confMap[queue.Name]
		if qc == nil {
			qc = &QueueConfig{
				QueueName: queue.Name,
				Password:  queue.Password,
				Topic:     map[string]*TopicConfig{},
			}
			confMap[queue.Name] = qc
		}
		storage := storages[topic.Storage]
		if storage == nil {
			if stmp, err := (dao.Storage{}.Get(topic.Storage)); err != nil {
				return "", err
			} else {
				storage = stmp
				storages[topic.Storage] = storage
			}
		}
		retry := new(TimeRetry)
		err := json.Unmarshal([]byte(topic.TopicConfig), retry)
		if err != nil {
			retry.IsRetry = false
			retry.RetryTime = 0
		}
		qc.Topic[topic.Name] = &TopicConfig{
			Name:           topic.Name,
			Password:       topic.Password,
			RetryTimes:     topic.RetryTimes,
			MaxQueueLength: topic.MaxQueueLength,
			RunType:        topic.RunType,
			NumOfWorkers:   uint64(topic.NumOfWorkers),
			ScriptEntry:    topic.ConsumeFile,
			Storage: &StorageConfig{
				Type:           storage.Type,
				Address:        fmt.Sprintf("%s:%d", storage.Host, storage.Port),
				Auth:           storage.Password,
				MaxConnNum:     storage.MaxConnNum,
				MaxIdleNum:     storage.MaxIdleNum,
				MaxIdleSeconds: storage.MaxIdleSeconds,
			},
			CgiConfigKey: topic.CgiConfig,
			HttpConfig:   topic.HttpConfig,
			IsRetry:      retry.IsRetry,
			RetryTime:    retry.RetryTime,
		}
	}

	v, err := json.Marshal(confMap)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func (c ConfigController) getQueueConfig() (string, error) {
	confMap := make(map[string]*QueueConfig)
	queues, err := dao.Queue{}.Query(&dao.Query{})
	if err != nil {
		return "", err
	}
	for _, queue := range queues {
		topicConf, err := c.getTopicConfig(queue.Name)
		if err != nil {
			return "", err
		} else if len(topicConf) == 0 {
			continue
		}
		confMap[queue.Name] = &QueueConfig{
			QueueName: queue.Name,
			Password:  queue.Password,
			Topic:     topicConf,
		}
	}
	v, err := json.Marshal(confMap)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func (c ConfigController) getTagsConfig() (string, error) {
	tags := make(map[string]map[string][]string)
	topics, err := dao.Topic{}.Query(&dao.Query{})
	if err != nil {
		return "", err
	}
	for _, topic := range topics {
		if topic.Status != dao.TopicStatusEnable {
			continue
		}
		if queue, ok := tags[topic.System]; !ok {
			tags[topic.System] = map[string][]string{
				topic.Queue: {topic.Name},
			}
		} else if tps, ok := queue[topic.Queue]; !ok {
			tags[topic.System][topic.Queue] = []string{
				topic.Name,
			}
		} else {
			tags[topic.System][topic.Queue] = append(tps, topic.Name)
		}
	}
	v, err := json.Marshal(tags)
	if err != nil {
		return "", err
	}
	return string(v), nil
}

func (c ConfigController) getMachineConfig(hosts string) (map[string]*MachineConfig, error) {
	ms, err := dao.Machine{}.Query(&dao.Query{Keyword: hosts})
	if err != nil {
		return nil, err
	}
	confs := make(map[string]*MachineConfig)
	for _, m := range ms {
		if config, ok := confs[m.IP]; !ok {
			confs[m.IP] = &MachineConfig{
				SharedId: m.ID,
				Tags:     []string{m.System},
			}
		} else {
			config.Tags = append(config.Tags, m.System)
		}
	}
	// if host is deleted
	for _, host := range strings.Split(hosts, "\n") {
		if _, ok := confs[host]; !ok {
			confs[host] = &MachineConfig{
				SharedId: 0,
				Tags:     []string{},
			}
		}
	}
	return confs, nil
}
