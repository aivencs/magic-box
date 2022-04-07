// 此包支持本地配置文件、远端配置中心的读取与解析
package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

// 使用枚举限定配置选择
type SupportType uint

const (
	Consul SupportType = iota + 1 // Consul 配置中心
)

const (
	DEFAULT_WATCH_INTERVAL = 180                     // 默认的更新检查间隔
	DEFAULT_HOST_CONSUL    = "http://localhost:8500" // Consul 服务默认地址
)

// 定义全局配置对象
var conf Conf
var once sync.Once

// 抽象接口
type Conf interface {
	PeriodicUpdate(ctx context.Context, option Option) // 定期更新
}

// 配置初始化时所用参数
type Option struct {
	Auth        bool        `json:"auth" label:"是否鉴权" desc:"鉴权时启用Username和Password"`
	Host        string      `json:"host" label:"路径" desc:"文件则填写文件路径" validate:"required"`
	Application string      `json:"application" label:"应用名称" desc:"必须与远端配置名称相同" validate:"required"`
	Env         string      `json:"env" label:"环境" desc:"推荐不同环境不同配置" validate:"required"`
	Type        string      `json:"type" label:"类型" desc:"用于指定配置格式类型，例如yaml/json" validate:"required"`
	Bind        interface{} `json:"bind" label:"用于映射配置的结构体" desc:"远端配置会被映射到结构体" validate:"required"`
	Username    string      `json:"username" label:"用户名" desc:"需要鉴权时使用"`
	Password    string      `json:"password" label:"密码" desc:"需要鉴权时使用"`
	Update      bool        `json:"update" label:"是否自动更新配置" desc:"默认不自动更新"`
	Interval    int         `json:"interval" label:"即时更新检查间隔" desc:"默认三分钟"`
}

// 初始化配置对象
func InitConf(ctx context.Context, name SupportType, option Option) error {
	c := conf
	var err error
	once.Do(func() {
		c = ConfFactory(ctx, name, option)
		if c == nil {
			err = errors.New("初始化失败")
		}
		conf = c
	})
	return err
}

// 配置的抽象工厂
func ConfFactory(ctx context.Context, name SupportType, option Option) Conf {
	switch name {
	case Consul:
		return NewConsulConf(ctx, option)
	default:
		return NewConsulConf(ctx, option)
	}
}

// Conf 结构体
// 基于 Consul
type ConsulConf struct {
	Kernel *viper.Viper
}

// 创建基于Consul的配置对象
func NewConsulConf(ctx context.Context, option Option) Conf {
	// 根据参数调整
	if utf8.RuneCountInString(option.Host) == 0 {
		option.Host = DEFAULT_HOST_CONSUL
	}
	if option.Auth {
		os.Setenv("CONSUL_HTTP_AUTH", fmt.Sprintf("%s:%s", option.Username, option.Password))
	}
	// 开始构建
	vip := viper.New()
	name := fmt.Sprintf("%s/%s", option.Application, option.Env)
	vip.SetConfigType(option.Type)
	vip.AddRemoteProvider("consul", option.Host, name)
	// 获取远端配置并映射到结构体
	err := vip.ReadRemoteConfig()
	if err != nil {
		return nil
	}
	if option.Bind != nil {
		vip.Unmarshal(&option.Bind)
	}
	return &ConsulConf{Kernel: vip}
}

// 定期更新
func (c *ConsulConf) PeriodicUpdate(ctx context.Context, option Option) {
	if option.Interval == 0 {
		option.Interval = DEFAULT_WATCH_INTERVAL
	}
	ticker := time.NewTicker(time.Second * time.Duration(option.Interval))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.Kernel.WatchRemoteConfig()
			if option.Bind != nil {
				c.Kernel.Unmarshal(&option.Bind)
			}
		default:
			continue
		}
	}
}
