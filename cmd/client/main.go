package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

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
		log.Fatalf("error creating client connection %s", err)
	}

	defer conn.Close()
	fmt.Println("\nConnection Sucessful")
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("error creating client channel: %s", err)
	}

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
		handlerMove(ch, gamestate),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Sub to move queue successful.")

	err = pubsub.SubscribeJSON(conn,
		routing.ExchangePerilTopic,
		routing.WarRecognitionsPrefix,
		routing.WarRecognitionsPrefix+".*",
		pubsub.Durable,
		handlerWar(ch, gamestate),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Sub to war queue successful.")

	fmt.Println("Launching game loop...")

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
			err = pubsub.PublishJSON(ch,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+userName,
				mv,
			)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Move message successfully published.")
		case "status":
			gamestate.CommandStatus()
		case "spam":
			if len(input) < 2 {
				fmt.Println("Usage: [spam] <number of times>")
				continue
			}
			amount, err := strconv.Atoi(input[1])
			if err != nil {
				fmt.Println("amount must a be a whole number")
				continue
			}
			for range amount {
				msg := gamelogic.GetMaliciousLog()
				err = pubsub.PublishGOB(
					ch,
					routing.ExchangePerilTopic,
					routing.GameLogSlug+"."+userName,
					routing.GameLog{
						CurrentTime: time.Now(),
						Message:     msg,
						Username:    userName,
					},
				)
				if err != nil {
					fmt.Printf("error pusblshing spam to logs: %s", err)
					break
				}
			}
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
