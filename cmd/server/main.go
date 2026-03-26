package main

import (
	"fmt"
	"log"

	"github.com/OminousOmelet/learn-pub-sub-starter/internal/gamelogic"
	"github.com/OminousOmelet/learn-pub-sub-starter/internal/pubsub"
	"github.com/OminousOmelet/learn-pub-sub-starter/internal/routing"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	const connStr string = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp091.Dial(connStr)
	if err != nil {
		log.Fatalf("error creating server connection: %s", err)
	}

	defer conn.Close()
	fmt.Println("Connection Sucessful")
	connCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("error creating server channel: %s", err)
	}

	gamelogic.PrintServerHelp()

	err = pubsub.SubscribeGOB(conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.Durable,
		handlerLogs(),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Server successfully subbed to logs.")

	for {
		userInput := gamelogic.GetInput()
		if len(userInput) == 0 {
			continue
		}
		switch userInput[0] {
		case "pause":
			fmt.Println("Sending 'pause' message...")
			err = pubsub.PublishJSON(connCh,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: true})
			if err != nil {
				fmt.Println(err)
			}
		case "resume":
			fmt.Println("Sending 'resume' message...")
			err = pubsub.PublishJSON(connCh,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: false})
			if err != nil {
				fmt.Println(err)
			}
		case "quit":
			fmt.Println("exiting...")
			return
		case "help":
			gamelogic.PrintServerHelp()
		default:
			fmt.Println("unrecognized command")
		}
	}
}
