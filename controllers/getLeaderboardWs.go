package controllers

import (
	"fmt"
	"go-trivia/shared"
	"go-trivia/utils"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var leaderboardConnections = make(map[*websocket.Conn]bool)
var leaderboardLock = &sync.Mutex{}

func GetLeaderboardWs(c echo.Context) error {
	conn, err := shared.Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		fmt.Println("Error upgrading to leaderboard websocket")
		return err
	}
	defer func(conn *websocket.Conn) {
		leaderboardLock.Lock()
		defer leaderboardLock.Unlock()
		delete(leaderboardConnections, conn)
		conn.Close()
	}(conn)

	leaderboardLock.Lock()
	leaderboardConnections[conn] = true
	leaderboardLock.Unlock()

	// send initial list of players
	fmt.Println("sending initial leaderboard")
	utils.SendLeaderboard(conn)

	// send when new data is available
	for _ = range shared.LeaderboardChan {
		fmt.Println("LeaderboardChan update")
		go broadcastLeaderboard()
	}

	return nil
}

func broadcastLeaderboard() {
	leaderboardLock.Lock()
	defer leaderboardLock.Unlock()

	for conn, _ := range leaderboardConnections {
		go utils.SendLeaderboard(conn)
	}
}
