package utils

import (
	"fmt"
	"go-trivia/shared"
	"sort"

	"github.com/gorilla/websocket"
)

func SendBuzzedIn(conn *websocket.Conn) {
	fmt.Println("SendBuzzedIn")

	// list of all players
	players := PlayersList()

	shared.Lock.RLock()
	defer shared.Lock.RUnlock()

	// sort players by buzz in time, then alphabetically
	sort.Slice(players, func(i, j int) bool {
		if players[i].BuzzIn.Before(players[j].BuzzIn) {
			return true
		} else if players[i].BuzzIn.After(players[j].BuzzIn) {
			return false
		} else {
			return players[i].Name < players[j].Name
		}
	})

	// list of all players and their buzz in times
	playersWithBuzzIn := make([][]string, 0)
	for _, player := range players {
		if player.BuzzIn.IsZero() {
			continue
		}
		playersWithBuzzIn = append(playersWithBuzzIn, []string{player.Name, player.BuzzIn.Format("03:04:05.000 PM")})
	}

	err := conn.WriteJSON(playersWithBuzzIn)
	if err != nil {
		fmt.Println("Error writing to websocket", err)
	}
}
