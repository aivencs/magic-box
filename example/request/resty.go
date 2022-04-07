package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aivencs/magic-box/pkg/request"
)

func main() {
	ctx := context.WithValue(context.Background(), "trace", "r001")
	for i := 0; i < 2; i++ {
		request.InitRequest(ctx, "resty", request.Option{})
		res, err := request.Get(ctx, request.Param{
			Link:    "https://www.example.com",
			Timeout: 5,
		})
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("status: ", res.StatusCode, "err: ", err)
	}
}
