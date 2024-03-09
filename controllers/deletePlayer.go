package controllers

import (
	"fmt"
	"go-trivia/shared"
	"go-trivia/utils"

	"github.com/labstack/echo/v4"
)

func DeletePlayer(c echo.Context) error {
	realPlayer, correctPassword, _, err := utils.ParseJSON(c)
	if err != nil {
		fmt.Println("Error decoding json")
		return c.String(400, "Bad Request: Invalid JSON")
	}

	// verify password
	if !correctPassword {
		fmt.Println("Unauthorized")
		return c.String(401, "Unauthorized")
	}

	shared.Lock.Lock()
	defer shared.Lock.Unlock()

	// verify player exists
	playerName := utils.CleanName(realPlayer)
	if _, ok := shared.PlayerData[playerName]; !ok {
		fmt.Println("Player not found")
		return c.String(400, "Bad Request: Player not found")
	}

	// delete player
	delete(shared.PlayerData, playerName)

	shared.LeaderboardChan <- true

	return c.String(200, "Player deleted")
}
