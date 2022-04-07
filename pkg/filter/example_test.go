package filter

import (
	"context"
	"fmt"
)

func ExampleBloomFilter() {
	ctx := context.WithValue(context.Background(), "trace", "ctx-filter-001")
	opt := Option{
		Host:     "localhost:6379",
		Auth:     true,
		Username: "",
		Password: "password",
		Database: "",
		Table:    "",
		DB:       1,
		Key:      "seeds",
	}
	InitFilter(ctx, BLOOM_FILTER, opt)
	payload := "19619c9e08f0ed4cc147e211efa8c3fb"
	res, err := Add(ctx, payload)
	fmt.Println(res, err) // output: false nil
	ex, err := Exist(ctx, payload)
	fmt.Println(ex, err) // output: true nil
}
