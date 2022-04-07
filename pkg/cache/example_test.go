package cache

import (
	"context"
	"fmt"
)

func ExampleRedisCache() {
	ctx := context.WithValue(context.Background(), "trace", "ctx-cache-001")
	opt := Option{
		Host:     "localhost:6379",
		Auth:     true,
		Username: "",
		Password: "password",
		Database: "",
		Table:    "",
		DB:       1,
	}
	InitCache(ctx, REDIS, opt)
	payload := "19619c9e08f0ed4cc147e211efa8c3f0"
	r, err := cache.SetEx(ctx, payload, 1, 20)
	fmt.Println(r, err) // output: OK nil
	ov := cache.Overdue(ctx, payload)
	fmt.Println(ov)                             // output: true
	fmt.Println(cache.Set(ctx, payload, "105")) // output: OK nil
	val, err := cache.Get(ctx, payload)
	fmt.Println(string(val.([]uint8)), err) // output: 105 <nil>
}
