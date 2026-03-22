package main

import (
	"fmt"

	"github.com/OminousOmelet/learn-pub-sub-starter/internal/gamelogic"
	"github.com/OminousOmelet/learn-pub-sub-starter/internal/pubsub"
	"github.com/OminousOmelet/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		// function returned as "handler", and 'ps' is filled when handler is called (feelsWierd)
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(mv gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(mv)
		var ack pubsub.AckType
		if outcome == gamelogic.MoveOutComeSafe || outcome == gamelogic.MoveOutcomeMakeWar {
			ack = pubsub.Ack
		} else {
			ack = pubsub.NackDiscard
		}
		return ack
	}
}
