package NacosClient

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"sync"
	"time"
)

var (
	C       *ConfigClient
	onceCon = &sync.Once{}
)

type ConfigClient struct {
	v *viper.Viper
}

func init() {
	startInfo()
	onceCon.Do(new)
}

// new
// @Author liangdongpo
// @Description 创建客户端
// @Date 10:34 下午 2021/11/14
// @Param
// @return
func new() {
	v := viper.New()
	v.SetConfigName(FN)
	v.SetConfigType(Ext)
	v.AddConfigPath(I.Path)
	err := v.ReadInConfig()
	if err != nil {
		log.Printf(err.Error())
		return
	}
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		if I.ShowChangeLog != "hidden" {
			log.Printf("Config file change: %s op: %d\n", in.Name, in.Op)
		}
	})
	C = &ConfigClient{v: v}
}

// Get
// @Author liangdongpo
// @Description 获取某个缓存，只支持默认参数下
// @Date 2:25 下午 2021/11/18
// @Param
// @return
func (g *ConfigClient) Get(key string) interface{} {
	return g.v.Get(key)
}

// GetString
// @Author liangdongpo
// @Description 获取string类型的缓存，只支持默认参数下
// @Date 2:30 下午 2021/11/18
// @Param
// @return
func (g *ConfigClient) GetString(key string) string {
	return g.v.GetString(key)
}

// GetBool 获取bool 类型 支持默认参数
func (g *ConfigClient) GetBool(key string) bool {
	return g.v.GetBool(key)
}

// GetInt 获取int 类型 支持默认参数
func (g *ConfigClient) GetInt(key string) int {
	return g.v.GetInt(key)
}

func (g *ConfigClient) GetInt32(key string) int32 {
	return g.v.GetInt32(key)
}

func (g *ConfigClient) GetInt64(key string) int64 {
	return g.v.GetInt64(key)
}

func (g *ConfigClient) GetUint(key string) uint {
	return g.v.GetUint(key)
}

func (g *ConfigClient) GetUint32(key string) uint32 {
	return g.v.GetUint32(key)
}

func (g *ConfigClient) GetUint64(key string) uint64 {
	return g.v.GetUint64(key)
}

func (g *ConfigClient) GetFloat64(key string) float64 {
	return g.v.GetFloat64(key)
}

func (g *ConfigClient) GetTime(key string) time.Time {
	return g.v.GetTime(key)
}

func (g *ConfigClient) GetDuration(key string) time.Duration {
	return g.v.GetDuration(key)
}

func (g *ConfigClient) GetIntSlice(key string) []int {
	return g.v.GetIntSlice(key)
}

func (g *ConfigClient) GetStringSlice(key string) []string {
	return g.v.GetStringSlice(key)
}

func (g *ConfigClient) GetStringMap(key string) map[string]interface{} {
	return g.v.GetStringMap(key)
}

func (g *ConfigClient) GetStringMapString(key string) map[string]string {
	return g.v.GetStringMapString(key)
}

func (g *ConfigClient) GetStringMapStringSlice(key string) map[string][]string {
	return g.v.GetStringMapStringSlice(key)
}

func (g *ConfigClient) GetSizeInBytes(key string) uint {
	return g.v.GetSizeInBytes(key)
}

func (g *ConfigClient) AllSettings() map[string]interface{} {
	return g.v.AllSettings()
}
