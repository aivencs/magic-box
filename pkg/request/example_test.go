package request

import (
	"context"
	"fmt"
)

func ExampleGet() {
	ctx := context.WithValue(context.Background(), "trace", "r001")
	for i := 0; i < 2; i++ {
		InitRequest(ctx, "resty", Option{})
		res, err := request.Get(ctx, Param{
			Link:    "https://www.example.com",
			Method:  GET,
			Timeout: 5,
		})
		fmt.Println("status: ", res.StatusCode, "err: ", err)
	}
}

func ExamplePost() {
	ctx := context.WithValue(context.Background(), "trace", "r002")
	for i := 0; i < 2; i++ {
		InitRequest(ctx, "resty", Option{})
		res, err := request.Get(ctx, Param{
			Link:    "https://www.example.com",
			Method:  POST,
			Timeout: 5,
		})
		fmt.Println("status: ", res.StatusCode, "err: ", err)
	}
}
