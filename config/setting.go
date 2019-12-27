package config

import (
	"flag"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	TokenSign     = "pepper_bus"
	DefaultAvatar = "https://avatars0.githubusercontent.com/u/58962967?s=60&v=4"
)

type LDAP struct {
	Enable  bool   `yaml:"enable"`
	Address string `yaml:"address"`
	BaseDN  string `yaml:"base-dn"`
}

type Keeper struct {
	DiscoverAddress string `yaml:"discover"`
	AdminAddress    string `yaml:"admin-addr"`
	Domain          string `yaml:"domain"`
}

type Base struct {
	Listen        int    `yaml:"listen"`
	AdminPassword string `yaml:"admin-pass"`
}

type Config struct {
	Dashboard Base     `yaml:"dashboard"`
	Keeper    Keeper   `yaml:"gokeeper"`
	LDAP      LDAP     `yaml:"ldap"`
	Postgres  Database `yaml:"postgres"`
	Redis     Redis    `yaml:"redis"`
	ETCD      ETCD     `yaml:"etcd"`
}

var (
	FilePath string
	Mode     string
)

var (
	GlobalConfig *Config
	HttpClient   = &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Second,
			}).DialContext,
			IdleConnTimeout:     10 * time.Second,
			MaxIdleConnsPerHost: 10,
		},
		Timeout: 5 * time.Second,
	}
)

func Init() {
	flag.StringVar(&FilePath, "cfg", "", "yaml config file path")
	flag.StringVar(&Mode, "e", "", "environment(release, test, debug)")

	flag.Parse()

	cfg, err := ioutil.ReadFile(FilePath)
	if err != nil {
		panic(err)
	}
	if err = yaml.Unmarshal(cfg, GlobalConfig); err != nil {
		panic(err)
	}

	InitDBs()
}
