package controllers

import (
	"fmt"
	"go-trivia/shared"
	"go-trivia/utils"

	"github.com/labstack/echo/v4"
)

func GetLeaderboardWs(c echo.Context) error {
	conn, err := shared.Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		fmt.Println("Error upgrading to leaderboard websocket")
		return err
	}
	defer conn.Close()

	// send initial list of players
	fmt.Println("sending initial leaderboard")
	utils.SendLeaderboard(conn)

	// send when new data is available
	for _ = range shared.LeaderboardChan {
		fmt.Println("LeaderboardChan update")
		go utils.SendLeaderboard(conn)
	}

	return nil
}
