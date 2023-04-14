package nacosconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 一些常量配置
// @Author 东坡
// @Description 具体参配置说明
// @Date 11:17 AM 2023/4/12
// @Param NacosApiHost 客户端不需要设置，如果需要更新配置内容，需要配置服务的api接口地址
// @return
const (
	FN            = "application" //生成配置文件的名称
	path          = "HOME"        //生成配置文件的根目录
	pathDir       = "go-ns-config"
	NacosApiHost  = "NACOS_API_HOST"                 //nacos api 服务地址
	ServerHost    = "GO_NS_CONFIG_SERVER_HOST"       //服务端地址
	showChangeLog = "GO_GRPC_CONFIG_SHOW_CHANGE_LOG" //是否显示文件修改提示
	Username      = "NACOS_USERNAME"                 //用户名
	Password      = "NACOS_PASSWORD"                 //密码
)

type Client struct {
	NamespaceId   string
	DataId        string
	Group         string
	Path          string
	ServerHost    string
	ShowChangeLog string
	ConfigType    string //配置的类型，目前支持：json、yaml
	Username      string //用户名
	Password      string //密码

	v  *viper.Viper
	ch chan interface{}
}

// NewClient
// @Author 东坡
// @Description 初始化客户端
// @Date 7:04 PM 2023/4/11
// @Param
// @return
func NewClient(namespaceId, dataId, group, configType string) (*Client, error) {
	client := &Client{}
	if namespaceId == "" {
		return nil, errors.New("NamespaceId cannot be empty")
	}
	if dataId == "" {
		return nil, errors.New("DataId cannot be empty")
	}
	if group == "" {
		return nil, errors.New("group cannot be empty")
	}
	ext := make(map[string]string)
	ext = map[string]string{"json": "json", "yaml": "yaml"}

	//后缀不符合要求
	if _, ok := ext[configType]; !ok {
		return nil, errors.New("configType err")
	}
	client.NamespaceId = namespaceId
	client.DataId = dataId
	client.Group = group
	client.ShowChangeLog = os.Getenv(showChangeLog)
	client.ConfigType = configType
	if os.Getenv(ServerHost) == "" {
		return nil, errors.New("server host cannot be empty")
	}

	if os.Getenv(Username) != "" {
		client.Username = os.Getenv(Username)
	}

	if os.Getenv(Password) != "" {
		client.Password = os.Getenv(Username)
	}

	client.ServerHost = os.Getenv(ServerHost)

	client.Path = filepath.Join(os.Getenv(path), pathDir, client.NamespaceId, client.DataId, client.Group, client.ConfigType)
	err := os.MkdirAll(client.Path, 0755)
	if err != nil {
		return nil, err
	}
	client.ch = make(chan interface{})
	return client, err
}

// SetWatcher
// @Author 东坡
// @Description 设置监听
// @Date 7:05 PM 2023/4/11
// @Param
// @return
func (c *Client) SetWatcher() error {
	// 创建clientConfig
	clientConfig := constant.ClientConfig{
		NamespaceId:         c.NamespaceId, // 如果需要支持多namespace，我们可以场景多个client,它们有不同的NamespaceId。当namespace是public时，此处填空字符串。
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		RotateTime:          "1h",
		MaxAge:              3,
		LogLevel:            "debug",
		Username:            c.Username,
		Password:            c.Password,
	}

	// 至少一个ServerConfig
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      c.ServerHost,
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
		return err
	}
	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: c.DataId,
		Group:  c.Group})
	//先获取一次数据
	err = ioutil.WriteFile(filepath.Join(c.Path, fmt.Sprintf("%s.%s", FN, c.ConfigType)), []byte(content), 0644)

	if err != nil {
		log.Printf("GetConfig err: %v", err)
		return err
	}
	go func() {
		err = configClient.ListenConfig(vo.ConfigParam{
			DataId: c.DataId,
			Group:  c.Group,
			OnChange: func(namespace, group, dataId, data string) {
				//fmt.Println("group:" + group + ", dataId:" + dataId + ", data:" + data)
				_ = ioutil.WriteFile(filepath.Join(c.Path, fmt.Sprintf("%s.%s", FN, c.ConfigType)), []byte(data), 0644)
			},
		})
		//暂时一直卡在这里让程序一直运行
		<-c.ch
		fmt.Println("end")
	}()

	err = c.NewViper()
	if err != nil {
		return err
	}
	return nil
}

// AddConfigs
// @Author 东坡
// @Description 更新内容
// @Date 11:29 AM 2023/4/12
// @Param
// @return
func (c *Client) AddConfigs(data string) (bool, error) {
	if os.Getenv(NacosApiHost) == "" {
		return false, errors.New("server api host cannot be empty")
	}
	apiUrl := os.Getenv(NacosApiHost) + "/nacos/v1/cs/configs"
	postValue := url.Values{}
	postValue.Set("tenant", c.NamespaceId)
	postValue.Set("dataId", c.DataId)
	postValue.Set("group", c.Group)
	postValue.Set("content", data)
	postValue.Set("type", c.ConfigType)
	contentType := "application/x-www-form-urlencoded"
	//参数，多个用&隔开
	postData := strings.NewReader(postValue.Encode())

	resp, err := http.Post(apiUrl, contentType, postData)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	var res bool
	err = json.Unmarshal(body, &res)
	log.Println(res)
	if err != nil {
		return false, err
	}
	return res, nil
}

// NewViper
// @Author 东坡
// @Description 创建 Viper
// @Date 7:22 PM 2023/4/11
// @Param
// @return
func (c *Client) NewViper() error {
	v := viper.New()
	v.SetConfigName(FN)
	v.SetConfigType(c.ConfigType)
	v.AddConfigPath(c.Path)
	err := v.ReadInConfig()
	if err != nil {
		log.Printf(err.Error())
		return err
	}
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		log.Printf("Config file change: %s op: %d\n", in.Name, in.Op)
	})
	//赋值
	c.v = v
	return nil
}

// Get
// @Author  mail@liangdongpo.com
// @Description 获取某个缓存，只支持默认参数下
// @Date 2:25 下午 2023/4/11
// @Param
// @return
func (c *Client) Get(key string) interface{} {
	return c.v.Get(key)
}

// GetString
// @Author  mail@liangdongpo.com
// @Description 获取string类型的缓存，只支持默认参数下
// @Date 2:30 下午 2023/4/11
// @Param
// @return
func (c *Client) GetString(key string) string {
	return c.v.GetString(key)
}

// GetBool 获取bool 类型 支持默认参数
func (c *Client) GetBool(key string) bool {
	return c.v.GetBool(key)
}

// GetInt 获取int 类型 支持默认参数
func (c *Client) GetInt(key string) int {
	return c.v.GetInt(key)
}

func (c *Client) GetInt32(key string) int32 {
	return c.v.GetInt32(key)
}

func (c *Client) GetInt64(key string) int64 {
	return c.v.GetInt64(key)
}

func (c *Client) GetUint(key string) uint {
	return c.v.GetUint(key)
}

func (c *Client) GetUint32(key string) uint32 {
	return c.v.GetUint32(key)
}

func (c *Client) GetUint64(key string) uint64 {
	return c.v.GetUint64(key)
}

func (c *Client) GetFloat64(key string) float64 {
	return c.v.GetFloat64(key)
}

func (c *Client) GetTime(key string) time.Time {
	return c.v.GetTime(key)
}

func (c *Client) GetDuration(key string) time.Duration {
	return c.v.GetDuration(key)
}

func (c *Client) GetIntSlice(key string) []int {
	return c.v.GetIntSlice(key)
}

func (c *Client) GetStringSlice(key string) []string {
	return c.v.GetStringSlice(key)
}

func (c *Client) GetStringMap(key string) map[string]interface{} {
	return c.v.GetStringMap(key)
}

func (c *Client) GetStringMapString(key string) map[string]string {
	return c.v.GetStringMapString(key)
}

func (c *Client) GetStringMapStringSlice(key string) map[string][]string {
	return c.v.GetStringMapStringSlice(key)
}

func (c *Client) GetSizeInBytes(key string) uint {
	return c.v.GetSizeInBytes(key)
}

func (c *Client) AllSettings() map[string]interface{} {
	return c.v.AllSettings()
}
