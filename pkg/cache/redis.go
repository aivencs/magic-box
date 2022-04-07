package cache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/aivencs/magic-box/pkg/validate"
	redigo "github.com/gomodule/redigo/redis"
)

// 使用枚举限定选择
type SupportType string

const (
	REDIS SupportType = "redis"
	// 定义默认值
	DEFAULT_MAXIDLE      = 20
	DEFAULT_IDLE_TIMEOUT = 120 * time.Second
	DEFAULT_MAXACTIVE    = 100
)

// 定义全局配置对象
var cache Cache
var once sync.Once

func init() {
	ctx := context.WithValue(context.Background(), "trace", "init-for-config")
	validate.InitValidate(ctx, "validator", validate.Option{})
}

// 抽象接口
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}) (interface{}, error)
	Overdue(ctx context.Context, key interface{}) bool
	SetEx(ctx context.Context, key string, value interface{}, sec int) (interface{}, error)
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
}

// 初始化对象
func InitCache(ctx context.Context, name SupportType, option Option) error {
	var c = cache
	var err error
	message, err := validate.Work(ctx, option)
	if err != nil {
		return errors.New(message)
	}
	once.Do(func() {
		c = CacheFactory(ctx, name, option)
		if c == nil {
			err = errors.New("初始化失败")
		}
		cache = c
	})
	return err
}

// 抽象工厂
func CacheFactory(ctx context.Context, name SupportType, option Option) Cache {
	switch name {
	case REDIS:
		return NewRedisCache(ctx, option)
	default:
		return NewRedisCache(ctx, option)
	}
}

// 结构体
// 基于Redis
type RedisCache struct {
	Pool *redigo.Pool
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

// 创建基于Redis的对象
func NewRedisCache(ctx context.Context, option Option) Cache {
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
	return &RedisCache{
		Pool: pool,
	}
}

func (c *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	r := c.Pool.Get()
	defer r.Close()
	return r.Do("GET", key)
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}) (interface{}, error) {
	r := c.Pool.Get()
	defer r.Close()
	return r.Do("SET", key, value)
}

func (c *RedisCache) Overdue(ctx context.Context, key interface{}) bool {
	r := c.Pool.Get()
	defer r.Close()
	res, err := r.Do("TTL", key)
	if err != nil {
		return false
	}
	return res.(int64) > 1
}

func (c *RedisCache) SetEx(ctx context.Context, key string, value interface{}, sec int) (interface{}, error) {
	r := c.Pool.Get()
	defer r.Close()
	return r.Do("SETEX", key, sec, value)
}

func Get(ctx context.Context, key string) (interface{}, error) {
	return cache.Get(ctx, key)
}

func Set(ctx context.Context, key string, value interface{}) (interface{}, error) {
	return cache.Set(ctx, key, value)
}

func SetEx(ctx context.Context, key string, value interface{}, sec int) (interface{}, error) {
	return cache.SetEx(ctx, key, value, sec)
}

func Overdue(ctx context.Context, key interface{}) bool {
	return cache.Overdue(ctx, key)
}
