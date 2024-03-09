package controllers

import (
	"fmt"
	"go-trivia/shared"
	"go-trivia/utils"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var buzzedInConnections = make(map[*websocket.Conn]bool)
var buzzedInLock = &sync.Mutex{}

func GetBuzzedInWs(c echo.Context) error {
	conn, err := shared.Upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		fmt.Println("Error upgrading to buzzed-in websocket")
		return err
	}
	defer func(conn *websocket.Conn) {
		buzzedInLock.Lock()
		defer buzzedInLock.Unlock()
		delete(buzzedInConnections, conn)
		conn.Close()
	}(conn)

	buzzedInLock.Lock()
	buzzedInConnections[conn] = true
	buzzedInLock.Unlock()

	// send initial list of players
	fmt.Println("sending initial buzzed in")
	// utils.SendBuzzedIn(conn)
	broadcastBuzzedIn()

	// send when new data is available
	for _ = range shared.BuzzChan {
		fmt.Println("BuzzChan update")
		go broadcastBuzzedIn()
	}

	return nil
}

func broadcastBuzzedIn() {
	buzzedInLock.Lock()
	defer buzzedInLock.Unlock()

	for conn, _ := range buzzedInConnections {
		go utils.SendBuzzedIn(conn)
	}
}
