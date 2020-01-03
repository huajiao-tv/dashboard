package config

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	cronCfg "github.com/huajiao-tv/peppercron/config"
	"github.com/huajiao-tv/peppercron/logic"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"go.etcd.io/etcd/clientv3"
)

var (
	MySQL       *gorm.DB
	ETCDClient  *clientv3.Client
	RedisClient redis.Cmdable
)

type Database struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"database"`
	User     string `yaml:"username"`
	Password string `yaml:"password"`
}

type ETCD struct {
	EndPoints []string `yaml:"endpoints"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
}

func (m Database) String() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		m.User, m.Password, m.Host, m.Port, m.Name)
}

type Redis struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func (m Redis) Options() *redis.Options {
	return &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", m.Host, m.Port),
		Password:     m.Password,
		DB:           m.DB,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}

func InitDBs() {
	var err error
	timer := time.NewTimer(time.Minute * 2)
	defer timer.Stop()

	// MySQL
	for {
		select {
		case <-timer.C:
			panic(err)
		default:
		}

		MySQL, err = gorm.Open("mysql", GlobalConfig.MySQL.String())
		if err != nil {
			time.Sleep(time.Second * 10)
			continue
		}

		break
	}

	MySQL.LogMode(true)
	MySQL.Set("gorm:table_options", "charset=utf8")
	MySQL.DB().SetConnMaxLifetime(time.Second * 10)

	// Redis
	RedisClient = redis.NewClient(GlobalConfig.Redis.Options())
	// Etcd
	ETCDClient, err = clientv3.New(clientv3.Config{
		Endpoints:   GlobalConfig.ETCD.EndPoints,
		DialTimeout: 5 * time.Second,
		Username:    GlobalConfig.ETCD.Username,
		Password:    GlobalConfig.ETCD.Password,
	})
	if err != nil {
		panic(err)
	}

	// Set peppercron global conf
	cronGlobal := cronCfg.SettingGlobal{
		FrontPort: cronCfg.DefaultFrontPort,
		AdminPort: cronCfg.DefaultAdminPort,
		JobResultTimesStorage: &cronCfg.SettingJobResultStorage{
			Type:        "redis",
			Addr:        fmt.Sprintf("%s:%d", GlobalConfig.Redis.Host, GlobalConfig.Redis.Port),
			Auth:        GlobalConfig.Redis.Password,
			MaxConnNum:  50,
			IdleTimeout: time.Second * 30,
		},
		JobResultStorage: &cronCfg.SettingJobResultStorage{
			Type:     "mysql",
			Addr:     fmt.Sprintf("%s:%d", GlobalConfig.MySQL.Host, GlobalConfig.MySQL.Port),
			Auth:     GlobalConfig.MySQL.Password,
			User:     GlobalConfig.MySQL.User,
			Database: GlobalConfig.MySQL.Name,
		},
	}
	cfgVal, _ := json.Marshal(&cronGlobal)
	_, err = ETCDClient.Put(context.TODO(), logic.GlobalConfig, string(cfgVal))
	if err != nil {
		panic(err)
	}
}
