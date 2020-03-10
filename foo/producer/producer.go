package main

import (
	"log"

	"github.com/krancour/brignext/pkg/messaging"

	redisMessaging "github.com/krancour/brignext/pkg/messaging/redis"
	myRedis "github.com/krancour/brignext/pkg/redis"
)

func main() {

	redisClient, err := myRedis.Client()
	if err != nil {
		log.Fatal(err)
	}

	producer := redisMessaging.NewProducer("foo", redisClient, nil)

	for i := 0; i < 1000; i++ {
		if err := producer.Publish(
			messaging.NewMessage([]byte("foo")),
		); err != nil {
			log.Fatal(err)
		}
	}
}
