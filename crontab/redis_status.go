package crontab

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/huajiao-tv/dashboard/dao"
)

type RedisStatus struct {
	*dao.Storage

	RunStatus string
	Msg       string
}

type redisClient struct {
	*dao.Storage
	redis.Cmdable
}

type RedisStateCollect struct {
	mu sync.RWMutex

	clients map[string]*redisClient
	states  map[string]*RedisStatus
}

func newRedisStateCollect() *RedisStateCollect {
	return &RedisStateCollect{
		clients: make(map[string]*redisClient),
		states:  make(map[string]*RedisStatus),
	}
}

func (s *RedisStateCollect) GetAll() []*RedisStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	states := make([]*RedisStatus, 0, len(s.states))
	for _, state := range s.states {
		states = append(states, state)
	}
	return states
}

func (s *RedisStateCollect) Update() {
	storages, err := dao.Storage{}.FindAll()
	if err != nil {
		log.Println("get storage list failed:", err)
		return
	}

	s.mu.Lock()
	for _, storage := range storages {
		key := fmt.Sprintf("%v-%v:%v:%v", storage.System, storage.Host, storage.Port, storage.Password)
		if _, ok := s.clients[key]; !ok {
			cli := redis.NewClient(&redis.Options{
				Addr:         fmt.Sprintf("%v:%v", storage.Host, storage.Port),
				Password:     storage.Password,
				DialTimeout:  time.Second * 3,
				ReadTimeout:  time.Second * 3,
				WriteTimeout: time.Second * 3,
			})
			s.clients[key] = &redisClient{
				Storage: storage,
				Cmdable: cli,
			}
		}
	}
	s.mu.Unlock()
}

func (s *RedisStateCollect) collect() {
	states := make(map[string]*RedisStatus, len(s.clients))

	for key, cli := range s.clients {
		info, err := cli.Info().Result()
		state := &RedisStatus{
			Storage:   cli.Storage,
			RunStatus: info,
		}
		if err != nil {
			state.Msg = err.Error()
		}
		states[key] = state
	}

	s.mu.Lock()
	s.states = states
	s.mu.Unlock()
}
