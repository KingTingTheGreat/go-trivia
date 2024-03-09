package controllers

import (
	"fmt"
	"go-trivia/shared"
	"go-trivia/utils"
	"sort"
	"strings"

	"github.com/labstack/echo/v4"
)

func PostStats(c echo.Context) error {
	// list of all players
	players := utils.PlayersList()

	shared.Lock.RLock()
	defer shared.Lock.RUnlock()

	// sort players by score, then by last update time
	sort.Slice(players, func(i, j int) bool {
		if players[i].Score == players[j].Score {
			return players[i].LastUpdate.Before(players[j].LastUpdate)
		}
		return players[i].Score > players[j].Score
	})

	// list of all players and their scores and correct answers
	playersWithStats := make([][]string, 0)
	for _, player := range players {
		playersWithStats = append(playersWithStats, []string{player.Name, fmt.Sprintf("%d", player.Score), strings.Trim(strings.Join(player.CorrectQuestions, ","), "[]")})
	}

	return c.JSON(200, playersWithStats)
}
