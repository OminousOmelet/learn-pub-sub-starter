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
		handlerPause(gamestate))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("\nGame state successfully prepared!")

	for {
		userInput := gamelogic.GetInput()
		if len(userInput) == 0 {
			continue
		}
		switch userInput[0] {
		case "spawn":
			err = gamestate.CommandSpawn(userInput)
			if err != nil {
				fmt.Println(err)
			}
		case "move":
			_, err := gamestate.CommandMove(userInput)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Unit successfully moved")
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
