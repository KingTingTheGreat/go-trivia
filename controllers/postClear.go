package controllers

import (
	"fmt"
	"go-trivia/shared"
	"go-trivia/utils"
	"time"

	"github.com/labstack/echo/v4"
)

func PostClear(c echo.Context) error {
	shared.Lock.Lock()
	defer shared.Lock.Unlock()

	_, correctPassword, _, err := utils.ParseJSON(c)
	if err != nil {
		fmt.Println("Error decoding json")
		return err
	}

	// verify password
	if !correctPassword {
		fmt.Println("Unauthorized")
		return c.String(401, "Unauthorized")
	}

	fmt.Println("Clear")
	// clear all player buzz in times
	for playerName, player := range shared.PlayerData {
		player.BuzzIn = time.Time{}
		shared.PlayerData[playerName] = player
	}

	shared.BuzzChan <- true

	return c.String(200, "Clear")
}
