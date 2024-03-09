package controllers

import (
	"go-trivia/shared"
	"go-trivia/utils"
	"sort"

	"github.com/labstack/echo/v4"
)

func GetPlayers(c echo.Context) error {
	shared.Lock.RLock()
	defer shared.Lock.RUnlock()
	players := make([]string, 0)
	for _, player := range shared.PlayerData {
		players = append(players, player.Name)
	}
	// sort alphabetically without case sensitivity
	sort.Slice(players, func(i, j int) bool {
		return utils.CleanName(players[i]) < utils.CleanName(players[j])
	})
	return c.JSON(200, players)
}
