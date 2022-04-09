package request

import (
	"context"
	"fmt"
	"log"

	"github.com/aivencs/magic-box/pkg/logger"
)

func ExampleGet() {
	ctx := context.WithValue(context.Background(), "trace", "r001")
	for i := 0; i < 2; i++ {
		initError := InitRequest(ctx, "resty", Option{
			LogType: logger.Zap,
			LogOption: logger.Option{
				Application: "resty",
				Env:         "product",
				Label:       "request",
				Encode:      logger.Json,
			},
		})
		if initError.Code != logger.SUCCESS {
			log.Fatal(initError.Label)
		}
		res, err := request.Get(ctx, Param{
			Link:    "https://www.example.com",
			Method:  GET,
			Timeout: 5,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("status: ", res.StatusCode, "err: ", err)
	}
}

func ExamplePost() {
	ctx := context.WithValue(context.Background(), "trace", "r002")
	for i := 0; i < 2; i++ {
		initError := InitRequest(ctx, "resty", Option{
			LogType: logger.Zap,
			LogOption: logger.Option{
				Application: "resty",
				Env:         "product",
				Label:       "request",
				Encode:      logger.Json,
			},
		})
		if initError.Code != logger.SUCCESS {
			log.Fatal(initError.Label)
		}
		res, err := request.Post(ctx, Param{
			Link:    "https://www.example.com",
			Method:  POST,
			Timeout: 5,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("status: ", res.StatusCode, "err: ", err)
	}
}
