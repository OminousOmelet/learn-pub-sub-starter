package main

import (
	"fmt"

	"github.com/OminousOmelet/learn-pub-sub-starter/internal/gamelogic"
	"github.com/OminousOmelet/learn-pub-sub-starter/internal/pubsub"
	"github.com/OminousOmelet/learn-pub-sub-starter/internal/routing"
	"github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		// function returned as "handler", and 'ps' is filled when handler is called (feelsWierd)
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(ch *amqp091.Channel, gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)
		var nack pubsub.AckType
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			nack = pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(ch,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.Player.Username,
				gamelogic.RecognitionOfWar{
					Attacker: move.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				fmt.Printf("error pubslishing war: %s", err)
				nack = pubsub.NackRequeue
			}
			nack = pubsub.Ack
		default:
			nack = pubsub.NackDiscard
		}
		return nack
	}
}

func handlerWar(gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, _, _ := gs.HandleWar(rw)
		var nack pubsub.AckType
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			nack = pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			nack = pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
		case gamelogic.WarOutcomeYouWon:
		case gamelogic.WarOutcomeDraw:
			nack = pubsub.Ack
		default:
			fmt.Printf("ERROR: unknown outcome: %v", outcome)
			nack = pubsub.NackDiscard
		}
		return nack
	}
}
