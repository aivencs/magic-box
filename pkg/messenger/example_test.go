package messenger

import (
	"context"
	"fmt"
	"time"
)

func ExampleRabbitMessenger() {
	ctx := context.WithValue(context.Background(), "trace", "ctx-messenger-001")
	option := Option{
		Host:      "localhost:5672",
		Auth:      true,
		Username:  "username",
		Password:  "password",
		Zone:      "/",
		Topic:     Topic{Product: "article_draft_eh", Consume: "article_draft"},
		Heartbeat: 120,
		Qos:       1,
	}
	InitMessenger(ctx, RABBIT, option)
	consumeObject, err := messenger.CreateConsume(ctx)
	if err != nil {
		fmt.Println(err)
	}
	cosm := consumeObject.(RabbitConsume)
	for {
		select {
		case carton := <-cosm.Consume:
			fmt.Println(map[string]interface{}{"p": carton.Priority, "m": string(carton.Body)})
			carton.Ack(false)
			messenger.Sent(ctx, SentPayload{
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
