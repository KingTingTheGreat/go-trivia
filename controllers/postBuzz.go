package controllers

import (
	"fmt"
	"go-trivia/shared"
	"go-trivia/types"
	"go-trivia/utils"
	"time"

	"github.com/labstack/echo/v4"
)

func PostBuzz(c echo.Context) error {
	fmt.Println("post buzz")

	shared.Lock.Lock()
	defer shared.Lock.Unlock()

	fmt.Println("post buzz got lock")

	realPlayer, _, _, err := utils.ParseJSON(c)
	if err != nil {
		fmt.Println("Error decoding json")
		return err
	}

	playerName := utils.CleanName(realPlayer)

	var player types.Player
	player, ok := shared.PlayerData[playerName]
	if ok {
		fmt.Println("existing player")
		// if existing player
		if player.BuzzIn.IsZero() {
			fmt.Println("first buzz in")
			// prevent buzzing in again
			player.BuzzIn = time.Now()
		}
	} else {
		fmt.Println("new player")
		// create new player if not exists
		player = types.Player{
			Name:             realPlayer,
			Score:            0,
			CorrectQuestions: make([]string, 0),
			LastUpdate:       time.Now(),
			BuzzIn:           time.Now(),
		}
		fmt.Println("created player")
		shared.LeaderboardChan <- true
		fmt.Println("sent data to leaderboardChan")
	}
	fmt.Println("setting player data...")
	shared.PlayerData[playerName] = player

	fmt.Println("sending data to BuzzChan")
	shared.BuzzChan <- true

	return c.String(200, fmt.Sprintf("%v", shared.QuestionNumber))
}
