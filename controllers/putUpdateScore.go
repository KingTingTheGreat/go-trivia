package controllers

import (
	"fmt"
	"go-trivia/shared"
	"go-trivia/utils"
	"time"

	"github.com/labstack/echo/v4"
)

func PutUpdateScore(c echo.Context) error {
	realPlayer, correctPassword, amountInt, err := utils.ParseJSON(c)
	if err != nil {
		fmt.Println("Error decoding json")
		return c.String(400, "Bad Request: Invalid JSON")
	}

	// verify password
	if !correctPassword {
		fmt.Println("Unauthorized")
		return c.String(401, "Unauthorized")
	}

	// verify playername and amount
	playerName := utils.CleanName(realPlayer)

	shared.Lock.Lock()
	defer shared.Lock.Unlock()

	player, ok := shared.PlayerData[playerName]
	if !ok {
		fmt.Println("Player not found")
		return c.String(400, "Bad Request: Player not found")
	}

	// update last update time
	player.LastUpdate = time.Now()

	// update player score
	if player.Score += int(amountInt); player.Score < 0 {
		player.Score = 0
	}

	// update player correct questions
	if amountInt > 0 {
		player.CorrectQuestions = append(player.CorrectQuestions, fmt.Sprintf("%d", shared.QuestionNumber))
	} else {
		player.CorrectQuestions = append(player.CorrectQuestions, fmt.Sprintf("-%d", shared.QuestionNumber))
	}

	shared.PlayerData[playerName] = player

	go func() { shared.LeaderboardChan <- true }()

	return c.String(200, fmt.Sprintf("%v", player.Score))
}
