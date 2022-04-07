package filter

import (
	"context"
	"errors"
	"sync"
	"time"

	redisbloom "github.com/RedisBloom/redisbloom-go"
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

// 抽象接口
type Filter interface {
	Exist(ctx context.Context, val string) (bool, error)
	Add(ctx context.Context, val string) (bool, error)
}

// 初始化时所用参数
type Option struct {
	Host        string
	Auth        bool
	Username    string
	Password    string
	Database    string
	Table       string
	DB          int
	MaxIdle     int
	IdleTimeout time.Duration
	MaxActive   int
	Key         string
}

// 初始化对象
func InitFilter(ctx context.Context, name SupportType, option Option) error {
	var c = filter
	var err error
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
		Wait:        true,
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
