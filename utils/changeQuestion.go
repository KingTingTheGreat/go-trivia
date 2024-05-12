package utils

import (
	"fmt"
	"go-trivia/shared"
	"time"

	"github.com/labstack/echo/v4"
)

func ChangeQuestion(c echo.Context, inc bool) error {
	_, correctPassword, _, err := ParseJSON(c)
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

	// increment or decrement question number
	if inc {
		shared.QuestionNumber += 1
	} else {
		shared.QuestionNumber -= 1
	}

	// clear all player buzz in times
	for playerName, player := range shared.PlayerData {
		player.BuzzIn = time.Time{}
		shared.PlayerData[playerName] = player
	}

	go func() { shared.BuzzChan <- true }()

	return c.String(200, fmt.Sprintf("%v", shared.QuestionNumber))
}
