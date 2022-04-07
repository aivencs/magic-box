// 日志包
package logger

import (
	"context"
	"errors"
	"os"
	"sync"

	"github.com/aivencs/magic-box/pkg/validate"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 使用枚举限定选择
type SupportType string

// 使用枚举限定选择
type EncoderType string

// 使用枚举限定选择
type LoggerLevel string

const (
	// SupportType
	Zap SupportType = "zap" // Zap 日志包
	// EncoderType
	Json    EncoderType = "json"    // Json Encoder
	Console EncoderType = "console" // Console Encoder
	// LoggerLevel
	DEBUG LoggerLevel = "debug"
	INFO  LoggerLevel = "info"
	WARN  LoggerLevel = "warn"
	ERROR LoggerLevel = "error"
	FATAL LoggerLevel = "fatal"
)

const DEFAULT_CODE = SUCCESS
const DEFAULT_LEVEL = INFO

// 定义全局配置对象
var logger Logger
var once sync.Once

func init() {
	ctx := context.WithValue(context.Background(), "trace", "init-for-logger")
	validate.InitValidate(ctx, "validator", validate.Option{})
}

// 抽象接口
type Logger interface {
	Debug(ctx context.Context, message Message)
	Info(ctx context.Context, message Message)
	Warn(ctx context.Context, message Message)
	Error(ctx context.Context, message Message)
	Fatal(ctx context.Context, message Message)
}

type Field = zapcore.Field
type Level = zapcore.Level
type Encoder = zapcore.Encoder
type EncoderConfig = zapcore.EncoderConfig

// 初始化时所用参数
type Option struct {
	Application string      `json:"application" label:"应用名称" desc:"必须与远端配置名称相同" validate:"required"`
	Env         string      `json:"env" label:"环境" desc:"推荐不同环境不同配置" validate:"required"`
	Label       string      `json:"label" label:"别称" desc:"用于后续日志细分" validate:"required"`
	Encode      EncoderType `json:"encoder" label:"输出格式" desc:""`
}

// 初始化对象
func InitLogger(ctx context.Context, name SupportType, option Option) error {
	c := logger
	var err error
	message, err := validate.Work(ctx, option)
	if err != nil {
		return errors.New(message)
	}
	once.Do(func() {
		c = LoggerFactory(ctx, name, option)
		if c == nil {
			err = errors.New("初始化失败")
		}
		logger = c
	})
	return err
}

// 抽象工厂
func LoggerFactory(ctx context.Context, name SupportType, option Option) Logger {
	switch name {
	case Zap:
		return NewZapLogger(ctx, option)
	default:
		return NewZapLogger(ctx, option)
	}
}

// 结构体
// 基于 Zap
type ZapLogger struct {
	Kernel      *zap.Logger
	Env         string `json:"env" label:"环境"`
	Application string `json:"application"  label:"应用名称"`
	Label       string `json:"label" label:"别称"`
}

// 日志信息主体
type Message struct {
	Text      string `json:"text" label:"描述文字"`
	Attr      Attr   `json:"attr" label:"其他信息"`
	Label     string `json:"label" label:"别称"`
	Remark    string `json:"remark" label:"备注"`
	Traceback string `json:"traceback" label:"错误栈信息"`
}

// 监控相关的日志字段
type Monitor struct {
	Final           bool        `json:"final" label:"是否为代码段日志"`
	Level           LoggerLevel `json:"level" label:"数据层面的日志级别"`
	Code            MessageCode `json:"code" label:"信息码"`
	ProcessDuration int64       `json:"process_duration" label:"运行耗时"`
	ProcessDelay    int64       `json:"process_delay" label:"延迟处理的时间"`
}

type Attr struct {
	Monitor Monitor                `json:"monitor" label:"监控相关的信息"`
	Inp     map[string]interface{} `json:"inp" label:"输入"`
	Oup     map[string]interface{} `json:"oup" label:"输出"`
}

// 创建基于Zap的日志对象
func NewZapLogger(ctx context.Context, option Option) Logger {
	enc := zapcore.EncoderConfig{
		TimeKey:        "when",
		LevelKey:       "level",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "traceback",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	// 构建输出方式
	zapWriterSync := zapcore.AddSync(os.Stdout)
	// 构建输出格式
	encoder := applyEncoder(option.Encode, enc)
	// 应用参数
	zapCore := zapcore.NewCore(
		encoder,
		zapWriterSync,
		zapcore.InfoLevel,
	)
	// 根据参数创建日志对象
	logger := zap.New(zapCore, zap.AddCaller(), zap.AddCallerSkip(3))
	defer logger.Sync()
	return &ZapLogger{
		Kernel:      logger,
		Env:         option.Env,
		Application: option.Application,
		Label:       option.Label,
	}
}

// 应用输出格式
func applyEncoder(types EncoderType, enc EncoderConfig) Encoder {
	switch types {
	case Json:
		return zapcore.NewJSONEncoder(enc)
	case Console:
		return zapcore.NewConsoleEncoder(enc)
	default:
		return zapcore.NewJSONEncoder(enc)
	}
}

func (c *ZapLogger) build(ctx context.Context, message Message) []zapcore.Field {
	if message.Attr.Monitor.Code == 0 {
		message.Attr.Monitor.Code = DEFAULT_CODE
	}
	if message.Attr.Monitor.Level == "" {
		message.Attr.Monitor.Level = DEFAULT_LEVEL
	}
	if message.Label == "" {
		message.Label = c.Label
	}
	content := map[string]zapcore.Field{
		"trace":       zap.String("trace", ctx.Value("trace").(string)),
		"env":         zap.String("env", c.Env),
		"application": zap.String("application", c.Application),
		"label":       zap.String("label", message.Label),
		"remark":      zap.String("remark", message.Remark),
		"traceback":   zap.String("traceback", message.Traceback),
		"attr":        zap.Any("attr", message.Attr),
	}
	result := []zapcore.Field{}
	for _, field := range content {
		result = append(result, field)
	}
	return result
}

func (c *ZapLogger) write(ctx context.Context, level string, message Message) {
	content := c.build(ctx, message)
	switch level {
	case string(DEBUG):
		c.Kernel.Info(message.Text, content...)
	case string(INFO):
		c.Kernel.Info(message.Text, content...)
	case string(WARN):
		c.Kernel.Warn(message.Text, content...)
	case string(ERROR):
		c.Kernel.Error(message.Text, content...)
	case string(FATAL):
		c.Kernel.Fatal(message.Text, content...)
	default:
		c.Kernel.Info(message.Text, content...)
	}

}

func (c *ZapLogger) Debug(ctx context.Context, message Message) {
	c.write(ctx, "debug", message)
}
func (c *ZapLogger) Info(ctx context.Context, message Message) {
	c.write(ctx, "info", message)
}
func (c *ZapLogger) Warn(ctx context.Context, message Message) {
	c.write(ctx, "warn", message)
}
func (c *ZapLogger) Error(ctx context.Context, message Message) {
	c.write(ctx, "error", message)
}
func (c *ZapLogger) Fatal(ctx context.Context, message Message) {
	c.write(ctx, "fatal", message)
}

func Debug(ctx context.Context, message Message) {
	logger.Debug(ctx, message)
}
func Info(ctx context.Context, message Message) {
	logger.Info(ctx, message)
}
func Warn(ctx context.Context, message Message) {
	logger.Warn(ctx, message)
}
func Error(ctx context.Context, message Message) {
	logger.Error(ctx, message)
}
func Fatal(ctx context.Context, message Message) {
	logger.Fatal(ctx, message)
}
