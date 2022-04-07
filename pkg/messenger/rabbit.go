package messenger

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aivencs/magic-box/pkg/validate"
	"github.com/streadway/amqp"
)

// 使用枚举限定选择
type SupportType string

const (
	RABBIT SupportType = "rabbitmq"
	// 定义默认值
	DEFAULT_QOS = 1
)

// 定义全局配置对象
var messenger Messenger

var once sync.Once

func init() {
	ctx := context.WithValue(context.Background(), "trace", "init-for-messenger")
	validate.InitValidate(ctx, "validator", validate.Option{})
}

// 抽象接口
type Messenger interface {
	CreateConsume(ctx context.Context) (interface{}, error)
	GetConnect() interface{}
	GetTopic() Topic
	Sent(ctx context.Context, payload SentPayload) error
}

// 初始化时所用参数
type Option struct {
	Host      string `json:"host" label:"服务地址" validate:"required"`
	Auth      bool   `json:"auth" label:"是否鉴权" desc:"默认不鉴权"`
	Zone      string `json:"zone" label:"操作区" validate:"required"`
	Username  string `json:"username" label:"用户名"`
	Password  string `json:"password" label:"密码"`
	Topic     Topic  `json:"topic" label:"topic" validate:"required"`
	Heartbeat int    `json:"heartbeat" label:"心跳间隔"`
	Qos       int    `json:"qos" label:"限流数"`
}

type SentPayload struct {
	Topic    string      `json:"topic" label:"生产主题" validate:"required"`
	Message  string      `json:"message" label:"消息" validate:"required"`
	Priority uint8       `json:"priority" label:"优先级" validate:"required"`
	Channel  interface{} `json:"channel" label:"信道" validate:"required"`
}

// 初始化对象
func InitMessenger(ctx context.Context, name SupportType, option Option) error {
	var c = messenger
	var err error
	message, err := validate.Work(ctx, option)
	if err != nil {
		return errors.New(message)
	}
	once.Do(func() {
		c = MessengerFactory(ctx, name, option)
		if c == nil {
			err = errors.New("初始化失败")
		}
		messenger = c
	})
	return err
}

// 抽象工厂
func MessengerFactory(ctx context.Context, name SupportType, option Option) Messenger {
	switch name {
	case RABBIT:
		return NewRabbitMessenger(ctx, option)
	default:
		return NewRabbitMessenger(ctx, option)
	}
}

// 结构体
// 基于Rabbitmq
type RabbitMessenger struct {
	Connect *amqp.Connection
	Topic   Topic
	Qos     int
	Channel *amqp.Channel
}

type Topic struct {
	Product string `json:"product" label:"生产主题" validate:"required"`
	Consume string `json:"consume" label:"消费主题" validate:"required"`
	Bad     string `json:"bad" label:"错误缓冲主题" validate:"required"`
	Forward string `json:"forward" label:"转发主题" validate:"required"`
}

type RabbitConsume struct {
	Channel *amqp.Channel
	Consume <-chan amqp.Delivery
}

// 创建基于Rabbitmq的对象
func NewRabbitMessenger(ctx context.Context, option Option) Messenger {
	conf := amqp.Config{
		Heartbeat: time.Second * time.Duration(option.Heartbeat),
	}
	address := fmt.Sprintf("amqp://%s%s", option.Host, option.Zone)
	if option.Auth {
		address = fmt.Sprintf("amqp://%s:%s@%s%s", option.Username, option.Password, option.Host, option.Zone)
	}
	conn, err := amqp.DialConfig(address, conf)
	if err != nil {
		return nil
	}
	if option.Qos == 0 {
		option.Qos = DEFAULT_QOS
	}
	return &RabbitMessenger{
		Topic:   option.Topic,
		Connect: conn,
		Qos:     option.Qos,
	}
}

func (c *RabbitMessenger) Sent(ctx context.Context, payload SentPayload) error {
	chl := payload.Channel.(*amqp.Channel)
	err := chl.Publish(
		c.Topic.Product, "", false, false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(payload.Message),
			Priority:    payload.Priority,
		},
	)
	return err
}

func (c *RabbitMessenger) GetTopic() Topic {
	return c.Topic
}

func (c *RabbitMessenger) GetConnect() interface{} {
	return c.Connect
}

func (c *RabbitMessenger) CreateConsume(ctx context.Context) (interface{}, error) {
	chl, err := c.Connect.Channel()
	if err != nil {
		return chl, err
	}
	err = chl.Qos(c.Qos, 0, false)
	if err != nil {
		return chl, err
	}
	consume, err := chl.Consume(
		c.Topic.Consume, // topic
		"",              // consumer
		false,           // autoAck
		false,           // exclusive
		false,           // noLocal
		false,           // noWait
		nil,             // args
	)
	return RabbitConsume{Consume: consume, Channel: chl}, err
}

func CreateConsume(ctx context.Context) (interface{}, error) {
	return messenger.CreateConsume(ctx)
}

func GetConnect() interface{} {
	return messenger.GetConnect()
}

func GetTopic() Topic {
	return messenger.GetTopic()
}

func Sent(ctx context.Context, payload SentPayload) error {
	return messenger.Sent(ctx, payload)
}
