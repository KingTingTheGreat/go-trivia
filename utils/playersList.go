package utils

import (
	"go-trivia/shared"
	"go-trivia/types"
)

func PlayersList() []types.Player {
	shared.Lock.Lock()
	defer shared.Lock.Unlock()

	players := make([]types.Player, 0)
	for _, player := range shared.PlayerData {
		players = append(players, player)
	}
	return players
}
