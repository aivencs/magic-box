package main

import (
	"context"

	"github.com/aivencs/magic-box/pkg/logger"
)

func main() {
	ctx := context.WithValue(context.Background(), "trace", "t00")
	// 初始化日志对象
	logger.InitLogger(ctx, logger.Zap, logger.Option{
		Application: "zap-log",
		Env:         "dev",
		Label:       "detail",
		Encode:      logger.Json})
	// 1
	ctx = context.WithValue(context.Background(), "trace", "t01")
	logger.Info(ctx, logger.Message{Text: "操作失败", Remark: "标题替代正文", Traceback: "按规则未找到正文",
		Attr: logger.Attr{
			Inp: map[string]interface{}{"link": "http://localhost:9087"},
			Oup: map[string]interface{}{"res": "title"},
			Monitor: logger.Monitor{
				Final:           true,
				Level:           logger.FATAL,
				Code:            logger.CHECK,
				ProcessDuration: 200,
				ProcessDelay:    20930,
			},
		},
	})
	// 2
	ctx = context.WithValue(context.Background(), "trace", "t02")
	logger.Error(ctx, logger.Message{Text: "work", Remark: "说明", Traceback: "调用超时", Label: "render",
		Attr: logger.Attr{
			Inp: map[string]interface{}{"application": "spanic-service-net"},
			Oup: map[string]interface{}{"result": ""},
			Monitor: logger.Monitor{
				Level:           logger.ERROR,
				Code:            logger.PVERROR,
				ProcessDuration: 5001,
				ProcessDelay:    2037,
			},
		},
	})
}