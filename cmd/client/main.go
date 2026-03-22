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
	fmt.Println("Starting Peril client...")
	const connStr string = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp091.Dial(connStr)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	fmt.Println("\nConnection Sucessful")

	userName, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatal(err)
	}

	gamestate := gamelogic.NewGameState(userName)

	err = pubsub.SubscribeJSON(conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+userName,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gamestate),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Sub to pause queue successful.")

	err = pubsub.SubscribeJSON(conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+userName,
		routing.ArmyMovesPrefix+".*",
		pubsub.Transient,
		handlerMove(gamestate),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Sub to move queue successful.")

	fmt.Println("Launching game loop...")
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Error starting publishing channel: %s", err)
	}
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
		case "spawn":
			err = gamestate.CommandSpawn(input)
			if err != nil {
				fmt.Println(err)
			}
		case "move":
			mv, err := gamestate.CommandMove(input)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = pubsub.PublishJSON(ch, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+userName, mv)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Move message successfully published.")
		case "status":
			gamestate.CommandStatus()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "help":
			gamelogic.PrintClientHelp()
		case "quit":
			fmt.Println("exiting...")
			return
		default:
			fmt.Println("unrecognized command")
		}
	}
}
