package request

import (
	"context"
	"crypto/tls"
	"errors"
	"net/url"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/go-resty/resty/v2"
)

// 使用枚举限定对象类型
type SupportType string

// 使用枚举限定请求方式
type MethodType string

const (
	RESTY SupportType = "resty"
	// method
	GET  MethodType = "GET"
	POST MethodType = "POST"
)

var once sync.Once

// 定义全局请求对象
var request Request

type Request interface {
	Get(ctx context.Context, param Param) (Result, error)
	Post(ctx context.Context, param Param) (Result, error)
}

// 初始化时所用参数
type Option struct {
}

// 结构体
// 基于Resty
type RestyRequest struct{}

// 请求参数
type Param struct {
	Link             string     `json:"link" label:"网址" validate:"required,url"`
	Method           MethodType `json:"method" label:"请求方式" validate:"required"`
	Payload          string     `json:"payload" label:"参数"`
	Timeout          int        `json:"timeout" label:"超时时间"`
	Proxy            string     `json:"proxy" label:"IP代理"`
	EnableSkipVerify bool       `json:"enable_skip_verify" label:"跳过证书" desc:"默认不开启"`
	EnableHeader     bool       `json:"enable_header" label:"根据网址设置请求头基本参数" desc:"默认不开启"`
}

// 请求结果
type Result struct {
	Text       string
	StatusCode int
	Response   interface{}
}

// 初始化对象
func InitRequest(ctx context.Context, name SupportType, option Option) error {
	c := request
	var err error
	once.Do(func() {
		c = RequestFactory(ctx, name, option)
		if c == nil {
			err = errors.New("初始化失败")
		}
		request = c
	})
	return err
}

// 抽象工厂
func RequestFactory(ctx context.Context, name SupportType, option Option) Request {
	switch name {
	case RESTY:
		return NewRestyRequest(ctx, option)
	default:
		return NewRestyRequest(ctx, option)
	}
}

// 创建基于Resty的请求对象
func NewRestyRequest(ctx context.Context, option Option) Request {
	return &RestyRequest{}
}

func (c *RestyRequest) Get(ctx context.Context, param Param) (Result, error) {
	return c.work(ctx, param)
}

func (c *RestyRequest) Post(ctx context.Context, param Param) (Result, error) {
	return c.work(ctx, param)
}

func (c *RestyRequest) work(ctx context.Context, param Param) (Result, error) {
	var result Result
	var response *resty.Response
	var err error
	serviceSafeString, _ := url.Parse(param.Link)
	client := resty.New()
	// 设置追踪编码
	client.SetHeaders(map[string]string{"X-REQUEST-ID": ctx.Value("trace").(string)})
	// 参数项的应用
	if param.Timeout > 0 {
		client.SetTimeout(time.Duration(param.Timeout) * time.Second)
	}
	if param.EnableSkipVerify {
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}
	if param.EnableHeader {
		client.SetHeaders(map[string]string{
			"Trace-ID":   ctx.Value("trace").(string),
			"Host":       serviceSafeString.Host,
			"Referer":    serviceSafeString.Host,
			"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36",
		})
	}
	if utf8.RuneCountInString(param.Proxy) > 6 {
		client.SetProxy(param.Proxy)
	}
	// 发出请求
	switch param.Method {
	case GET:
		response, err = client.R().SetBody(param.Payload).Get(param.Link)
	case POST:
		client.SetHeaders(map[string]string{
			"Content-Type": "application/json",
		})
		response, err = client.R().SetBody(param.Payload).Post(param.Link)
	default:
		response, err = client.R().SetBody(param.Payload).Get(param.Link)
	}
	// 请求结果处理
	if err != nil {
		return result, err
	}
	// 状态码处理
	if response.RawResponse.StatusCode < 299 {
		switch response.RawResponse.StatusCode {
		case 429:
			err = errors.New("并发超限")
		case 404:
			err = errors.New("资源不存在")
		case 200:
			err = nil
		case 201:
			err = nil
		default:
			err = errors.New("非正常状态码")
		}
	}
	// 构造结果并返回
	return Result{
		Text:       response.String(),
		StatusCode: response.RawResponse.StatusCode,
		Response:   response,
	}, err
}

// 暴露给外部调用
func Get(ctx context.Context, param Param) (Result, error) {
	return request.Get(ctx, param)
}

// 暴露给外部调用
func Post(ctx context.Context, param Param) (Result, error) {
	return request.Get(ctx, param)
}
