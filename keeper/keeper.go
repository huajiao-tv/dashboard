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
	ConfigClient = NewClient(config.GlobalConfig.Keeper.AdminAddress)

	opt := discovery.WithDiscovery("dashboard-client", []string{config.GlobalConfig.Keeper.Domain})
	DiscoveryClient = discovery.New(config.GlobalConfig.Keeper.DiscoverAddress, opt)
	go DiscoveryClient.Work()
}

func GetNodeList() ([]string, error) {
	res, err := DiscoveryClient.GetService(config.GlobalConfig.Keeper.Domain)
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
	return DiscoveryClient.GetServiceAddr(config.GlobalConfig.Keeper.Domain, AdminSchemaName)
}
