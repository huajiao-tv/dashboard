package keeper

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/huajiao-tv/dashboard/config"
)

const (
	Start   = "start"
	Stop    = "stop"
	Restart = "restart"

	Partner        = "pepper_bus"
	InnerSecretKey = "5ff10ecc78ada17c37b96fdf1ecb0c9e"
)

const (
	GetConfig    = "get"
	AddConfig    = "add"
	UpdateConfig = "update"

	DefaultSection = "DEFAULT"
)

type Client struct {
	keeper string
}

func NewClient(keeper string) *Client {
	return &Client{
		keeper: keeper,
	}
}

func (c *Client) Keeper() string {
	return c.keeper
}

// 查询 keeper 下所有集群
func (c *Client) QueryClusters() (Clusters, error) {
	resp := &QueryClustersResp{}
	args := url.Values{}
	if err := c.request("/domain/list", args, false, resp); err != nil {
		return nil, err
	} else {
		cs := resp.Data
		sort.Sort(cs)
		return cs, nil
	}
}

func (c *Client) QueryNodeList(cluster, component string) ([]*NodeInfo, error) {
	resp := &QueryNodeListResp{}
	args := url.Values{
		"domain":    []string{cluster},
		"component": []string{component},
	}
	if err := c.request("/node/list", args, false, resp); err != nil {
		return nil, err
	} else {
		return resp.Data, nil
	}
}

func (c *Client) QueryConfig(cluster string) ([]*FileConfig, error) {
	resp := &QueryConfigResp{}
	args := url.Values{
		"domain": []string{cluster},
	}
	if err := c.request("/conf/list", args, false, resp); err != nil {
		return nil, err
	} else {
		return resp.Data, nil
	}
}

func (c *Client) ReloadConfig(_ string) error {
	path := "/conf/reload"
	_ = path
	return nil
}

func (c *Client) RollbackConfig(_ string, _ int64) error {
	path := "/conf/rollback"
	_ = path
	return nil
}

func (c *Client) ManageConfig(cluster string, ops []*ConfigOperate, comment string) error {
	data, err := json.Marshal(ops)
	if err != nil {
		return err
	}
	args := url.Values{
		"domain":   []string{cluster},
		"operates": []string{string(data)},
		"note":     []string{comment},
	}
	resp := &BaseResponse{}
	if err := c.request("/conf/manage", args, true, resp); err != nil {
		return err
	} else {
		return nil
	}
}

func (c *Client) QueryHistory() error {
	path := "/package/list"
	_ = path
	return nil
}

func (c *Client) QueryNodeStatus() error {
	path := "/conf/status"
	_ = path
	return nil
}

func (c *Client) request(path string, form url.Values, post bool, data interface{}) (err error) {
	random := strconv.FormatInt(rand.Int63(), 10)
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	sign := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%v_%v_%v%v", Partner, random, timestamp, InnerSecretKey))))
	params := url.Values{
		"partner": []string{Partner},
		"rand":    []string{random},
		"time":    []string{timestamp},
		"guid":    []string{sign},
	}
	var resp *http.Response
	if post {
		uri := fmt.Sprintf("http://%s%s?%s", c.keeper, path, params.Encode())
		resp, err = config.HttpClient.PostForm(uri, form)
	} else {
		uri := fmt.Sprintf("http://%s%s?%s&%s", c.keeper, path, params.Encode(), form.Encode())
		resp, err = config.HttpClient.Get(uri)
	}
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("error status:" + resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, data); err != nil {
		return err
	}
	return nil
}
