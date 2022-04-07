package logger

import (
	"context"
	"log"
)

func ExampleZapLogger() {
	ctx := context.WithValue(context.Background(), "trace", "t00")
	// 初始化日志对象
	err := InitLogger(ctx, Zap, Option{
		Application: "zap-log",
		Env:         "dev",
		Label:       "detail",
		Encode:      Json})
	if err != nil {
		log.Fatal(err)
	}
	// 1
	ctx = context.WithValue(context.Background(), "trace", "t01")
	logger.Info(ctx, Message{Text: "操作失败", Remark: "标题替代正文", Traceback: "按规则未找到正文",
		Attr: Attr{
			Inp: map[string]interface{}{"link": "http://localhost:9087"},
			Oup: map[string]interface{}{"res": "title"},
			Monitor: Monitor{
				Final:           true,
				Level:           FATAL,
				Code:            CHECK,
				ProcessDuration: 200,
				ProcessDelay:    20930,
			},
		},
	})
	// 2
	ctx = context.WithValue(context.Background(), "trace", "t02")
	logger.Error(ctx, Message{Text: "work", Remark: "说明", Traceback: "调用超时", Label: "render",
		Attr: Attr{
			Inp: map[string]interface{}{"application": "spanic-service-net"},
			Oup: map[string]interface{}{"result": ""},
			Monitor: Monitor{
				Level:           ERROR,
				Code:            PVERROR,
				ProcessDuration: 5001,
				ProcessDelay:    2037,
			},
		},
	})
}
