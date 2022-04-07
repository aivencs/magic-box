package main

import (
	"context"

	"github.com/aivencs/magic-box/pkg/server"
)

func main() {
	ctx := context.WithValue(context.Background(), "trace", "v001")
	server.InitServer(ctx, server.SERVER_ECHO, server.Option{Port: 9817, Host: "localhost"})
	server.Work()
}
