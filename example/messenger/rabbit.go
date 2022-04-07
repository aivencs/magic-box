package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aivencs/magic-box/pkg/messenger"
)

func main() {
	ctx := context.WithValue(context.Background(), "trace", "ctx-messenger-001")
	option := messenger.Option{
		Host:      "localhost:5672",
		Auth:      true,
		Username:  "username",
		Password:  "password",
		Zone:      "/",
		Topic:     messenger.Topic{Product: "article_draft_eh", Consume: "article_draft"},
		Heartbeat: 120,
		Qos:       1,
	}
	messenger.InitMessenger(ctx, messenger.RABBIT, option)
	consumeObject, err := messenger.CreateConsume(ctx)
	if err != nil {
		fmt.Println(err)
	}
	cosm := consumeObject.(messenger.RabbitConsume)
	for {
		select {
		case carton := <-cosm.Consume:
			fmt.Println(map[string]interface{}{"p": carton.Priority, "m": string(carton.Body)})
			carton.Ack(false)
			messenger.Sent(ctx, messenger.SentPayload{
				Topic:    messenger.GetTopic().Product,
				Message:  fmt.Sprintf("abc-%s", string(carton.Body)),
				Priority: 5,
				Channel:  cosm.Channel,
			})
			time.Sleep(time.Second * 5)
		default:
			continue
		}
	}
}
