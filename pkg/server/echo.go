package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/aivencs/magic-box/pkg/kit"
	"github.com/aivencs/magic-box/pkg/logger"
	"github.com/aivencs/magic-box/pkg/request"
	"github.com/aivencs/magic-box/pkg/validate"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// 使用枚举限定选择
type SupportType string

type MethodType string

const (
	SERVER_ECHO SupportType = "echo"
	// 限定请求方式
	GET    MethodType = "GET"
	POST   MethodType = "POST"
	DELETE MethodType = "DELETE"
	PUT    MethodType = "PUT"
	// 定义默认值
	DEFAULT_HOST = ":"
)

var server Server
var once sync.Once
var routerLabel = map[string]string{}

func init() {
	ctx := context.WithValue(context.Background(), "trace", "init-for-server")
	validate.InitValidate(ctx, "validator", validate.Option{})
}

type Server interface {
	Work()
	AddRouter(payload RouterPayload, h echo.HandlerFunc, m ...echo.MiddlewareFunc)
	GetRouterLabel(path string) string
}

type EchoServer struct {
	Kernel *echo.Echo
	Port   int
	Host   string
}

type Option struct {
	Host               string `json:"host" label:"服务地址" desc:"默认为开放访问"`
	Port               int    `json:"port" label:"端口号" validate:"required,min=3000,max=10000"`
	DisableMiddCors    bool   `json:"disable_midd_cors" label:"cors中间件开关" desc:"默认开启"`
	DisableMiddRecover bool   `json:"disable_midd_recover" label:"recover中间件开关" desc:"默认开启"`
}

func InitServer(ctx context.Context, name SupportType, option Option) error {
	c := server
	message, err := validate.Work(ctx, option)
	if err != nil {
		return errors.New(message)
	}
	once.Do(func() {
		c = ServerFactory(ctx, name, option)
		if c == nil {
			err = errors.New("初始化失败")
		}
		server = c
	})
	return err
}

func ServerFactory(ctx context.Context, name SupportType, option Option) Server {
	switch name {
	case SERVER_ECHO:
		return NewEchoServer(ctx, option)
	default:
		return NewEchoServer(ctx, option)
	}
}

func NewEchoServer(ctx context.Context, option Option) Server {
	svr := echo.New()
	if !option.DisableMiddCors {
		svr.Use(middleware.CORS())
	}
	if !option.DisableMiddRecover {
		svr.Use(middleware.Recover())
	}
	if len(option.Host) == 0 {
		option.Host = DEFAULT_HOST
	} else {
		option.Host = kit.JoinString(option.Host, DEFAULT_HOST)
	}
	return &EchoServer{
		Kernel: svr,
		Port:   option.Port,
		Host:   option.Host,
	}
}

func (c *EchoServer) Work() {
	port := strconv.Itoa(c.Port)
	c.Kernel.Logger.Fatal(c.Kernel.Start(kit.JoinString(c.Host, port)))
}

type RouterPayload struct {
	Method MethodType
	Path   string
	Label  string
}

func (c *EchoServer) AddRouter(payload RouterPayload, h echo.HandlerFunc, m ...echo.MiddlewareFunc) {
	switch payload.Method {
	case GET:
		c.Kernel.GET(payload.Path, h, m...)
	case POST:
		c.Kernel.POST(payload.Path, h, m...)
	case DELETE:
		c.Kernel.DELETE(payload.Path, h, m...)
	case PUT:
		c.Kernel.PUT(payload.Path, h, m...)
	default:
		c.Kernel.GET(payload.Path, h, m...)
	}
	routerLabel[payload.Path] = payload.Label
}

func (c *EchoServer) GetRouterLabel(path string) string {
	return routerLabel[path]
}

func Work() {
	server.Work()
}

func AddRouter(payload RouterPayload, h echo.HandlerFunc, m ...echo.MiddlewareFunc) {
	server.AddRouter(payload, h, m...)
}

func GetRouterLabel(path string) string {
	return server.GetRouterLabel(path)
}

type bodyDumpResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

// 请求结果
type Result struct {
	Text       string
	StatusCode int
	Response   interface{}
	ErrorCode  logger.ErrorCode
}

func EmptyHandler(c echo.Context) error {
	res := Result{
		Text:       c.Get("message").(string),
		StatusCode: 200,
		Response:   nil,
		ErrorCode:  logger.GetErc(logger.PVERROR, ""),
	}
	return c.JSONPretty(http.StatusOK, res, "")
}

type Header struct {
	X_REQUEST_ID string `json:"X-REQUEST-ID" label:"追踪编码" validate:"required,min=16"`
}

// 日志中间件
func LoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) (err error) {
		var header Header
		label := server.GetRouterLabel(c.Request().URL.Path)
		// 获取并校验追踪编码
		ids := c.Request().Header.Values("X-REQUEST-ID")
		if len(ids) > 0 {
			header.X_REQUEST_ID = ids[0]
		}
		message, err := validate.Work(context.Background(), &header)
		if err != nil {
			c.Set("message", message)
			EmptyHandler(c)
			return
		}
		// 设置框架的Context
		c.Set("trace", header.X_REQUEST_ID)
		c.Set("label", label)
		// 创建新的 Context
		ctx := context.WithValue(context.Background(), "trace", header.X_REQUEST_ID)
		ctx = context.WithValue(ctx, "label", label)
		// 拦截响应
		responseBuffer := new(bytes.Buffer)
		mw := io.MultiWriter(c.Response().Writer, responseBuffer)
		writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: c.Response().Writer}
		response := request.Result{}
		c.Response().Writer = writer
		// 函数调用
		startT := time.Now()
		err = next(c)
		duration := time.Since(startT).Milliseconds()
		// 解析响应内容
		json.Unmarshal(responseBuffer.Bytes(), &response)
		// 构建日志信息
		logger.Info(ctx, logger.Message{
			Text:  response.ErrorCode.Label,
			Label: label,
			Attr: logger.Attr{
				Monitor: logger.Monitor{
					Final:           true,
					Level:           response.ErrorCode.Level,
					ProcessDuration: duration,
				},
				Inp: map[string]interface{}{
					"host":       c.Request().Host,
					"path":       c.Path(),
					"user-agent": c.Request().UserAgent(),
					"method":     c.Request().Method,
					"param":      c.Get("request"),
				},
				Oup: map[string]interface{}{
					"status_code": c.Response().Status,
				},
			},
		})
		return
	}
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
