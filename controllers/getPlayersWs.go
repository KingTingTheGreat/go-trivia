package controllers

import (
	"fmt"
	"go-trivia/shared"
	"go-trivia/utils"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var playersConnections = make(map[*websocket.Conn]bool)
var playersLock = &sync.Mutex{}	

func GetPlayersWs(c echo.Context) error {
	conn, err := shared.Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		fmt.Println("Error upgrading to players websocket")
		return err
	}
	defer func (conn *websocket.Conn) {
		playersLock.Lock()
		defer playersLock.Unlock()
		delete(leaderboardConnections,conn)
		conn.Close()
	}(conn)

	playersLock.Lock()
	playersConnections[conn] = true
	playersLock.Unlock()

	// send initial list of players
	fmt.Println("sending initial players list")
	go utils.SendPlayers(conn)

	// send when new data is available
	for {
		select {
			case <- shared.PlayersChan: 
				fmt.Println("PlayersChan update")
				go broadcastPlayers()
		}
	}
}

func broadcastPlayers() {
	playersLock.Lock()
	defer playersLock.Unlock()

	for conn := range playersConnections {
		go utils.SendPlayers(conn)
	}
}
