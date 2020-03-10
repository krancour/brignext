package main

import (
	"context"
	"log"
	"time"

	"github.com/krancour/brignext/pkg/messaging"
	redisMessaging "github.com/krancour/brignext/pkg/messaging/redis"
	myRedis "github.com/krancour/brignext/pkg/redis"
	"github.com/krancour/brignext/pkg/signals"
)

func main() {

	redisClient, err := myRedis.Client()
	if err != nil {
		log.Fatal(err)
	}

	consumer, err := redisMessaging.NewConsumer(
		redisClient,
		"foo",
		nil,
		handleMessage,
	)
	if err != nil {
		log.Fatal(err)
	}

	if err := consumer.Run(signals.Context()); err != nil {
		log.Fatal(err)
	}
}

func handleMessage(
	ctx context.Context,
	message messaging.Message,
) error {
	select {
	case <-time.After(time.Second):
	case <-ctx.Done():
		return ctx.Err()
	}
	log.Println(string(message.ID()))
	return nil
}
