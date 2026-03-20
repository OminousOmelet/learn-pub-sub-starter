package main

import (
	"fmt"

	"github.com/OminousOmelet/learn-pub-sub-starter/internal/gamelogic"
	"github.com/OminousOmelet/learn-pub-sub-starter/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		defer fmt.Print("> ")
		// function returned as "handler", and 'ps' is filled when handler is called (feelsWierd)
		gs.HandlePause(ps)
	}
}
