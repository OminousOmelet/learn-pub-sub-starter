package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/OminousOmelet/learn-pub-sub-starter/internal/pubsub"
	"github.com/OminousOmelet/learn-pub-sub-starter/internal/routing"
	"github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	const connStr string = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp091.Dial(connStr)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()
	fmt.Println("Connection Sucessful")

	// create new channel on the connection
	connCh, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}

	err = pubsub.PublishJSON(connCh, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
	if err != nil {
		log.Fatal(err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	fmt.Println("\nshutting down")
	os.Exit(0)
}
