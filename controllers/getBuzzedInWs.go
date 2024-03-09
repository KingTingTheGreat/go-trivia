package controllers

import (
	"fmt"
	"go-trivia/shared"
	"go-trivia/utils"

	"github.com/labstack/echo/v4"
)

func GetBuzzedInWs(c echo.Context) error {
	conn, err := shared.Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		fmt.Println("Error upgrading to buzzed-in websocket")
		return err
	}
	defer conn.Close()

	// send initial list of players
	fmt.Println("sending initial buzzed in")
	utils.SendBuzzedIn(conn)

	// send when new data is available
	for _ = range shared.BuzzChan {
		fmt.Println("BuzzChan update")
		go utils.SendBuzzedIn(conn)
	}

	return nil
}
