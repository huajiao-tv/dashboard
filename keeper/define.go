package keeper

type BaseResponse struct {
	Code    int    `json:"error_code"`
	Message string `json:"error"`
}

type ClusterInfo struct {
	Name    string `json:"name"`
	Version int64  `json:"version"`
}

type Clusters []*ClusterInfo

func (s Clusters) Len() int           { return len(s) }
func (s Clusters) Less(i, j int) bool { return s[i].Name < s[j].Name }
func (s Clusters) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type QueryClustersResp struct {
	BaseResponse
	Data Clusters `json:"data"`
}

type KeyConfig struct {
	Type     string `json:"type"`
	RawKey   string `json:"raw_key"`
	RawValue string `json:"raw_value"`
	Key      string `json:"key"`
}

type SectionConfig struct {
	Name string                `json:"name"`
	Keys map[string]*KeyConfig `json:"keys"`
}

type FileConfig struct {
	Name     string           `json:"name"`
	Sections []*SectionConfig `json:"sections"`
}

type QueryConfigResp struct {
	BaseResponse
	Data []*FileConfig `json:"data"`
}

type NodeConfig struct {
	Name    string                `json:"name"`
	Version int64                 `json:"version"`
	Data    map[string]*KeyConfig `json:"data"`
}

type ProcessBaseInfo struct {
	Pid       string `json:"Pid"`
	ParentPid string `json:"PPid"`
	Command   string `json:"Command"`
	State     string `json:"State"`
	StartTime string `json:"StartTime"`
}

type ProcessCpuInfo struct {
	UTime     int64  `json:"Utime"`
	STime     int64  `json:"Stime"`
	Cutime    int64  `json:"Cutime"`
	Cstime    int64  `json:"Cstime"`
	StartTime int64  `json:"StartTime"`
	LastUS    int64  `json:"LastUS"`
	LastTimer string `json:"LastTimer"`
	CpuUsage  string `json:"CpuUsage"`
	Pid       string `json:"Pid"`
	PPid      string `json:"PPid"`
	Command   string `json:"Command"`
	State     string `json:"State"`
}

type ProcessMemoryInfo struct {
	VmSize    int64  `json:"VmSize"`
	VmRss     int64  `json:"VmRss"`
	VmData    int64  `json:"VmData"`
	VmStk     int64  `json:"VmStk"`
	VmExe     int64  `json:"VmExe"`
	VmLib     int64  `json:"VmLib"`
	Pid       string `json:"Pid"`
	PPid      string `json:"PPid"`
	Command   string `json:"Command"`
	State     string `json:"State"`
	StartTime string `json:"StartTime"`
}

type ProcessInfo struct {
	Base *ProcessBaseInfo   `json:"Base"`
	Cpu  *ProcessCpuInfo    `json:"Cpu"`
	Mem  *ProcessMemoryInfo `json:"Mem"`
}

type CompileInfo struct {
	Operator  string `json:"operator"`
	TimeStamp string `json:"timestamp"`
	VCS       string `json:"vcs"`
	Version   string `json:"version"`
}

type NodeInfo struct {
	Id              string        `json:"id"`
	KeeperAddr      string        `json:"keeper_addr"`
	Cluster         string        `json:"domain"`
	Component       string        `json:"component"`
	Hostname        string        `json:"hostname"`
	StartTime       int64         `json:"start_time"`
	UpdateTime      int64         `json:"update_time"`
	RawSubscription []string      `json:"raw_subscription"`
	Status          int           `json:"status"`
	Version         int64         `json:"version"`
	CompileInfo     *CompileInfo  `json:"component_tags"`
	Subscription    []string      `json:"subscription"`
	Configs         []*NodeConfig `json:"struct_datas"`
	ProcessInfo     *ProcessInfo  `json:"proc"`
}

type QueryNodeListResp struct {
	BaseResponse
	Data []*NodeInfo `json:"data"`
}

type ConfigOperate struct {
	Action  string `json:"opcode"`
	Cluster string `json:"domain"`
	File    string `json:"file"`
	Section string `json:"section"`
	Key     string `json:"key"`
	Type    string `json:"type"`
	Value   string `json:"value"`
	Comment string `json:"note"`
	ID      int    `json:"id"`
}
