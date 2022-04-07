package filter

import (
	"context"
	"errors"
	"sync"
	"time"

	redisbloom "github.com/RedisBloom/redisbloom-go"
	"github.com/aivencs/magic-box/pkg/validate"
	redigo "github.com/gomodule/redigo/redis"
)

// 使用枚举限定选择
type SupportType string

const (
	BLOOM_FILTER SupportType = "bloom_filter"
	// 定义默认值
	DEFAULT_MAXIDLE      = 20
	DEFAULT_IDLE_TIMEOUT = 120 * time.Second
	DEFAULT_MAXACTIVE    = 100
)

// 定义全局配置对象
var filter Filter
var once sync.Once

func init() {
	ctx := context.WithValue(context.Background(), "trace", "init-for-filter")
	validate.InitValidate(ctx, "validator", validate.Option{})
}

// 抽象接口
type Filter interface {
	Exist(ctx context.Context, val string) (bool, error)
	Add(ctx context.Context, val string) (bool, error)
}

// 初始化时所用参数
type Option struct {
	Host        string        `json:"host" label:"服务地址" validate:"required"`
	Auth        bool          `json:"auth" label:"是否鉴权" desc:"默认不鉴权"`
	Username    string        `json:"username" label:"用户名"`
	Password    string        `json:"password" label:"密码"`
	Database    string        `json:"database" label:"数据库"`
	Table       string        `json:"table" label:"数据表"`
	DB          int           `json:"db" label:"数据库"`
	MaxIdle     int           `json:"max_idle" label:"最大空闲链接数"`
	IdleTimeout time.Duration `json:"idle_timeout" label:"空闲超时时间"`
	MaxActive   int           `json:"max_active" label:"最大链接数"`
	Key         string        `json:"key" label:"键名"`
}

// 初始化对象
func InitFilter(ctx context.Context, name SupportType, option Option) error {
	var c = filter
	var err error
	message, err := validate.Work(ctx, option)
	if err != nil {
		return errors.New(message)
	}
	once.Do(func() {
		c = FilterFactory(ctx, name, option)
		if c == nil {
			err = errors.New("初始化失败")
		}
		filter = c
	})
	return err
}

// 抽象工厂
func FilterFactory(ctx context.Context, name SupportType, option Option) Filter {
	switch name {
	case BLOOM_FILTER:
		return NewBloomFilter(ctx, option)
	default:
		return NewBloomFilter(ctx, option)
	}
}

// 结构体
// 基于
type BloomFilter struct {
	Kernel *redisbloom.Client
	Pool   *redigo.Pool
	Key    string
}

// 创建基于的对象
func NewBloomFilter(ctx context.Context, option Option) Filter {
	applyOption(option)
	pool := &redigo.Pool{
		MaxIdle:     option.MaxIdle,
		IdleTimeout: option.IdleTimeout,
		MaxActive:   option.MaxActive,
		Wait:        true, // 连接池无空闲连接时等待
		Dial: func() (redigo.Conn, error) {
			c, err := redigo.Dial("tcp", option.Host)
			if err != nil {
				return nil, err
			}
			if option.Auth {
				if _, err := c.Do("AUTH", option.Password); err != nil {
					c.Close()
					return nil, err
				}
				if _, err := c.Do("SELECT", option.DB); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redigo.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
	rbc := redisbloom.NewClientFromPool(pool, option.Key)
	return &BloomFilter{
		Kernel: rbc,
		Pool:   pool,
		Key:    option.Key,
	}
}

func applyOption(option Option) {
	if option.MaxIdle == 0 {
		option.MaxIdle = DEFAULT_MAXIDLE
	}
	if option.IdleTimeout == 0 {
		option.IdleTimeout = DEFAULT_IDLE_TIMEOUT
	}
	if option.MaxActive == 0 {
		option.MaxActive = DEFAULT_MAXACTIVE
	}
}

func (c *BloomFilter) Add(ctx context.Context, val string) (bool, error) {
	return c.Kernel.Add(c.Key, val)
}

func (c *BloomFilter) Exist(ctx context.Context, val string) (bool, error) {
	return c.Kernel.Exists(c.Key, val)
}

func Exist(ctx context.Context, val string) (bool, error) {
	return filter.Exist(ctx, val)
}

func Add(ctx context.Context, val string) (bool, error) {
	return filter.Add(ctx, val)
}
