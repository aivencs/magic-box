package main

import (
	"context"
	"log"

	"github.com/aivencs/magic-box/pkg/server"
)

func main() {
	ctx := context.WithValue(context.Background(), "trace", "v001")
	err := server.InitServer(ctx, server.SERVER_ECHO, server.Option{Port: 9817, Host: "localhost"})
	if err != nil {
		log.Fatal(err)
	}
	server.Work()
}
