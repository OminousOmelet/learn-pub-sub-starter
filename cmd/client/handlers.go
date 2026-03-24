package main

import (
	"fmt"
	"time"

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
				fmt.Printf("error pubslishing move message: %s", err)
				nack = pubsub.NackRequeue
			}
			nack = pubsub.Ack
		default:
			nack = pubsub.NackDiscard
		}

		return nack
	}
}

func handlerWar(ch *amqp091.Channel, gs *gamelogic.GameState) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := gs.HandleWar(rw)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			err := publishGameLog(ch,
				gs.GetUsername(),
				fmt.Sprintf("%s won a war against %s", winner, loser),
			)
			if err != nil {
				fmt.Printf("error pubslishing war message: %s", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			err := publishGameLog(ch,
				gs.GetUsername(),
				fmt.Sprintf("%s won a war against %s", winner, loser),
			)
			if err != nil {
				fmt.Printf("error pubslishing war message: %s", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			err := publishGameLog(ch,
				gs.GetUsername(),
				fmt.Sprintf("A war between %s and %s resulted in a draw.", winner, loser),
			)
			if err != nil {
				fmt.Printf("error pubslishing war message: %s", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack

		}

		fmt.Printf("ERROR: unknown outcome: %v", outcome)
		return pubsub.NackDiscard
	}
}

func publishGameLog(ch *amqp091.Channel, userName, msg string) error {
	err := pubsub.PublishGOB(ch,
		routing.ExchangePerilTopic,
		routing.GameLogSlug+"."+userName,
		routing.GameLog{
			CurrentTime: time.Now(),
			Message:     msg,
			Username:    userName,
		},
	)
	return err
}
