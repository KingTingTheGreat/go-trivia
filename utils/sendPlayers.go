package utils

import (
	"fmt"
	"go-trivia/shared"
	"sort"

	"github.com/gorilla/websocket"
)

func SendPlayers(conn *websocket.Conn) {
	players := PlayersList()

	shared.Lock.RLock()
	defer shared.Lock.RUnlock()

	// sort into alphabetical order
	sort.Slice(players, func(i, j int) bool {
		return players[i].Name < players[j].Name
	})

	playerNames := make([]string, len(players))
	for _, player := range players {
		playerNames = append(playerNames, player.Name)
	}

	err := conn.WriteJSON(playerNames)
	if err != nil {
		fmt.Println("Error writing to players websocket")
	}
}
