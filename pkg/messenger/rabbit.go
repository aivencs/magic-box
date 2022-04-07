package messenger

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

// 使用枚举限定选择
type SupportType string

const (
	RABBIT SupportType = "rabbitmq"
)

// 定义全局配置对象
var messenger Messenger

var once sync.Once

// 抽象接口
type Messenger interface {
	CreateConsume(ctx context.Context) (interface{}, error)
	GetConnect() interface{}
	GetTopic() Topic
	Sent(ctx context.Context, payload SentPayload) error
}

// 初始化时所用参数
type Option struct {
	Host      string
	Auth      bool
	Zone      string
	Username  string
	Password  string
	Topic     Topic
	Heartbeat int
	Active    bool
	Qos       int
}

type SentPayload struct {
	Topic    string
	Message  string
	Priority uint8
	Channel  interface{}
}

// 初始化对象
func InitMessenger(ctx context.Context, name SupportType, option Option) error {
	var c = messenger
	var err error
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
	Product string
	Consume string
	Bad     string
	Forward string
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
