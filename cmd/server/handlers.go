package main

import (
	"fmt"

	"github.com/OminousOmelet/learn-pub-sub-starter/internal/gamelogic"
	"github.com/OminousOmelet/learn-pub-sub-starter/internal/pubsub"
	"github.com/OminousOmelet/learn-pub-sub-starter/internal/routing"
)

func handlerLogs() func(routing.GameLog) pubsub.AckType {
	return func(gl routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")
		// function returned as "handler", and 'ps' is filled when handler is called (feelsWierd)
		err := gamelogic.WriteLog(gl)
		if err != nil {
			fmt.Printf("error writing logs: %s", err)
		}
		return pubsub.Ack
	}
}
