package utils

import (
	"fmt"
	"go-trivia/shared"
	"sort"

	"github.com/gorilla/websocket"
)

func SendLeaderboard(conn *websocket.Conn) {
	fmt.Println("sendLeaderboard")
	// list of all players
	players := PlayersList()

	shared.Lock.RLock()
	// defer shared.Lock.RUnlock()

	// sort players by score, then by last update time
	sort.Slice(players, func(i, j int) bool {
		if players[i].Score == players[j].Score {
			return players[i].LastUpdate.Before(players[j].LastUpdate)
		}
		return players[i].Score > players[j].Score
	})

	shared.Lock.RUnlock()

	// list of all players and their scores
	playersWithScores := make([][]string, 0)
	for _, player := range players {
		playersWithScores = append(playersWithScores, []string{player.Name, fmt.Sprintf("%d", player.Score)})
	}

	err := conn.WriteJSON(playersWithScores)
	if err != nil {
		fmt.Println("Error writing to leaderboard websocket")
	}
}
