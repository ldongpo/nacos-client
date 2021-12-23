package NacosClient

import (
	"errors"
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

const (
	FN            = "application" //生成配置文件的名称
	Ext           = "json"        //生成配置文件的后缀
	path          = "HOME"        //生成配置文件的根目录
	pathDir       = "go-ns-config"
	ServerHost    = "GO_NS_CONFIG_SERVER_HOST"       //服务端地址
	NamespaceId   = "GO_NS_CONFIG_NAMESPACE_ID"      //NamespaceId
	DataId        = "GO_NS_CONFIG_DATA_ID"           //DataId
	Group         = "GO_NS_CONFIG_GROUP"             //Group
	showChangeLog = "GO_GRPC_CONFIG_SHOW_CHANGE_LOG" //是否显示文件修改提示
)

var (
	I    *info
	once = &sync.Once{}
	ch   = make(chan interface{})
)

func startInfo() {
	once.Do(initInfo)
}

func initInfo() {
	if I != nil {
		return
	}
	var err error
	I, err = createInfo()
	if err != nil {
		log.Fatalf("Fatal error GRcpConfig newInfo : %v\n", err)
	}
	// 创建clientConfig
	clientConfig := constant.ClientConfig{
		NamespaceId:         I.NamespaceId, // 如果需要支持多namespace，我们可以场景多个client,它们有不同的NamespaceId。当namespace是public时，此处填空字符串。
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		RotateTime:          "1h",
		MaxAge:              3,
		LogLevel:            "debug",
	}

	// 至少一个ServerConfig
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      I.ServerHost,
			ContextPath: "/nacos",
			Port:        8848,
			Scheme:      "http",
		},
	}

	// 创建动态配置客户端的另一种方式 (推荐)
	configClient, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	if err != nil {
		log.Printf("configClient err: %v", err)
		return
	}
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: I.DataId,
		Group:  I.Group})
	//先获取一次数据
	err = ioutil.WriteFile(filepath.Join(I.Path, fmt.Sprintf("%s.%s", FN, Ext)), []byte(content), 0644)
	if err != nil {
		log.Printf("GetConfig err: %v", err)
		return
	}
	go func() {
		err = configClient.ListenConfig(vo.ConfigParam{
			DataId: I.DataId,
			Group:  I.Group,
			OnChange: func(namespace, group, dataId, data string) {
				//fmt.Println("group:" + group + ", dataId:" + dataId + ", data:" + data)
				_ = ioutil.WriteFile(filepath.Join(I.Path, fmt.Sprintf("%s.%s", FN, Ext)), []byte(data), 0644)
			},
		})
		//暂时一直卡在这里让程序一直运行
		<-ch
		fmt.Println("end")
	}()
}

func createInfo() (*info, error) {
	inf := &info{}
	if os.Getenv(NamespaceId) == "" {
		return nil, errors.New("NamespaceId cannot be empty")
	}
	if os.Getenv(DataId) == "" {
		return nil, errors.New("DataId cannot be empty")
	}
	if os.Getenv(Group) == "" {
		return nil, errors.New("group cannot be empty")
	}

	inf.NamespaceId = os.Getenv(NamespaceId)
	inf.DataId = os.Getenv(DataId)
	inf.Group = os.Getenv(Group)
	inf.ShowChangeLog = os.Getenv(showChangeLog)
	if os.Getenv(ServerHost) == "" {
		return nil, errors.New("server host cannot be empty")
	}
	inf.ServerHost = os.Getenv(ServerHost)

	inf.Path = filepath.Join(os.Getenv(path), pathDir, inf.NamespaceId, inf.DataId, inf.Group)
	err := os.MkdirAll(inf.Path, 0755)
	if err != nil {
		return inf, err
	}

	return inf, err
}

type info struct {
	NamespaceId   string
	DataId        string
	Group         string
	Path          string
	ServerHost    string
	ShowChangeLog string
}
