package server

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"github.com/aivencs/magic-box/pkg/kit"
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
var routerLabel map[string]string

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
