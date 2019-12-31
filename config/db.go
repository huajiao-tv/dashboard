package config

import (
	"fmt"
	"time"

	"github.com/etcd-io/etcd/clientv3"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
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
}
