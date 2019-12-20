package keeper

import (
	"github.com/huajiao-tv/dashboard/config"
	"github.com/huajiao-tv/gokeeper/client/discovery"
)

var (
	ConfigClient    *Client
	DiscoveryClient *discovery.Client
)

const (
	Component       = "main"
	AdminSchemaName = "admin"
)

func Init() {
	ConfigClient = NewClient(*config.KeeperAdmin)

	opt := discovery.WithDiscovery("dashboard-client", []string{*config.Domain})
	DiscoveryClient = discovery.New(*config.KeeperDiscovery, opt)
	go DiscoveryClient.Work()
}

func GetNodeList() ([]string, error) {
	res, err := DiscoveryClient.GetService(*config.Domain)
	if err != nil {
		return nil, err
	}
	if len(res.Instances) == 0 {
		return nil, nil
	}
	nodeList := make([]string, 0, len(res.Instances))
	for _, instances := range res.Instances {
		for _, instance := range instances {
			nodeList = append(nodeList, instance.Addrs[AdminSchemaName])
		}
	}
	return nodeList, err
}

func GetNode() (string, error) {
	return DiscoveryClient.GetServiceAddr(*config.Domain, AdminSchemaName)
}
