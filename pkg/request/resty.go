package request

import (
	"context"
	"crypto/tls"
	"errors"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/aivencs/magic-box/pkg/logger"
	"github.com/aivencs/magic-box/pkg/server"
	"github.com/aivencs/magic-box/pkg/validate"
	"github.com/go-resty/resty/v2"
)

// 使用枚举限定对象类型
type SupportType string

// 使用枚举限定请求方式
type MethodType string

const (
	RESTY SupportType = "resty"
	// 限定请求方式
	GET  MethodType = "GET"
	POST MethodType = "POST"
	// 定义默认值
	DEFAULT_TIEOUT = 10
)

var once sync.Once

// 定义全局请求对象
var request Request

func init() {
	ctx := context.WithValue(context.Background(), "trace", "init-for-request")
	validate.InitValidate(ctx, "validator", validate.Option{})
	logger.InitErrorCode()
	logger.InitLogger(ctx, logger.Zap, logger.Option{
		Application: "ac",
		Env:         "dev",
		Encode:      logger.Json,
		Label:       "request",
	})
}

type Request interface {
	Get(ctx context.Context, param Param) (server.Result, error)
	Post(ctx context.Context, param Param) (server.Result, error)
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
	Method           MethodType `json:"method" label:"请求方式"`
	Payload          string     `json:"payload" label:"参数"`
	Timeout          int        `json:"timeout" label:"超时时间"`
	Proxy            string     `json:"proxy" label:"IP代理"`
	EnableSkipVerify bool       `json:"enable_skip_verify" label:"跳过证书" desc:"默认不开启"`
	EnableHeader     bool       `json:"enable_header" label:"根据网址设置请求头基本参数" desc:"默认不开启"`
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

func (c *RestyRequest) Get(ctx context.Context, param Param) (server.Result, error) {
	return c.work(ctx, param)
}

func (c *RestyRequest) Post(ctx context.Context, param Param) (server.Result, error) {
	return c.work(ctx, param)
}

func (c *RestyRequest) work(ctx context.Context, param Param) (server.Result, error) {
	erc := logger.GetDefaultErc()
	var result server.Result
	var response *resty.Response
	var err error
	// 参数校验
	message, err := validate.Work(ctx, param)
	if err != nil {
		return server.Result{}, errors.New(message)
	}
	// 前期准备
	serviceSafeString, _ := url.Parse(param.Link)
	client := resty.New()
	// 设置追踪编码
	client.SetHeaders(map[string]string{"X-REQUEST-ID": ctx.Value("trace").(string)})
	// 参数项的应用
	if param.Timeout == 0 {
		param.Timeout = DEFAULT_TIEOUT
	}
	client.SetTimeout(time.Duration(param.Timeout) * time.Second)
	if param.EnableSkipVerify {
		client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}
	if param.EnableHeader {
		client.SetHeaders(map[string]string{
			"X-REQUEST-ID": ctx.Value("trace").(string),
			"Host":         serviceSafeString.Host,
			"Referer":      serviceSafeString.Host,
			"User-Agent":   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.159 Safari/537.36",
		})
	}
	if utf8.RuneCountInString(param.Proxy) > 6 {
		client.SetProxy(param.Proxy)
	}
	// 发出请求
	startT := time.Now()
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
	duration := time.Since(startT).Milliseconds()
	// 请求结果处理
	if err != nil {
		erc := logger.GetErc(logger.DVERROR, "请求时发生错误")
		if strings.Contains(err.Error(), "Client.Timeout ") {
			erc = logger.GetErc(logger.TIMEOUT, "")
		}
		logger.Info(ctx, logger.Message{
			Text:      erc.Label,
			Label:     ctx.Value("label").(string),
			Traceback: err.Error(),
			Attr: logger.Attr{
				Monitor: logger.Monitor{
					Level:           erc.Level,
					Code:            erc.Code,
					ProcessDuration: duration,
				},
			},
		})
		return result, errors.New(erc.Label)
	}
	// 状态码处理
	if response.RawResponse.StatusCode > 201 {
		switch response.RawResponse.StatusCode {
		case 429:
			erc = logger.GetErc(logger.LIMITERROR, "")
			err = errors.New(erc.Label)
		case 404:
			erc = logger.GetErc(logger.CHECK, "资源不存在")
			err = errors.New(erc.Label)
		case 200:
			err = nil
		case 201:
			err = nil
		default:
			erc = logger.GetErc(logger.STATUSERROR, "")
			err = errors.New(erc.Label)
		}
	}
	// 构造结果并返回
	return server.Result{
		Text:       response.String(),
		StatusCode: response.RawResponse.StatusCode,
		Response:   response,
		ErrorCode:  erc,
	}, err
}

// 暴露给外部调用
func Get(ctx context.Context, param Param) (server.Result, error) {
	param.Method = GET
	return request.Get(ctx, param)
}

// 暴露给外部调用
func Post(ctx context.Context, param Param) (server.Result, error) {
	param.Method = POST
	return request.Get(ctx, param)
}
