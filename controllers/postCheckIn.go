package controllers

import (
	"fmt"
	"go-trivia/shared"
	"go-trivia/types"
	"go-trivia/utils"
	"time"

	"github.com/labstack/echo/v4"
)

func PostCheckIn(c echo.Context) error {
	shared.Lock.Lock()
	defer shared.Lock.Unlock()

	realPlayer, _, _, err := utils.ParseJSON(c)
	if err != nil {
		fmt.Println("Error decoding json")
		return err
	}

	playerName := utils.CleanName(realPlayer)
	fmt.Printf("got player: %s\n", playerName)

	// create new player if not exists
	if _, ok := shared.PlayerData[playerName]; !ok {
		fmt.Println("creating new player")
		shared.PlayerData[playerName] = types.Player{
			Name:             realPlayer,
			Score:            0,
			CorrectQuestions: make([]string, 0),
			LastUpdate:       time.Now(),
			BuzzIn:           time.Time{},
		}
		fmt.Println("check in sending to leaderboard")
		shared.LeaderboardChan <- true
		shared.PlayersChan <- true
	}

	return c.String(200, fmt.Sprintf("%v", shared.QuestionNumber))
}
