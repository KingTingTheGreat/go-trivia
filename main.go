package main

import (
	"encoding/json"
	"fmt"
	"go-trivia/configs"
	"go-trivia/controllers"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type Player struct {
	Name             string
	Score            int
	CorrectQuestions []string
	LastUpdate       time.Time
	BuzzIn           time.Time
}

func main() {
	// prevent data races
	Lock := &sync.Mutex{}

	// websocket upgrader
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// player data
	playerData := make(map[string]Player)

	questionNumber := 0
	password := configs.Password()

	parseJSON := func(c echo.Context) (string, bool, int64, error) {
		var realPlayer string
		var correctPassword bool
		var amountInt int64

		json_map := make(map[string]interface{})
		err := json.NewDecoder(c.Request().Body).Decode(&json_map)
		if err != nil {
			fmt.Println("Error decoding json")
			return realPlayer, correctPassword, amountInt, err
		}

		// name
		realPlayer, _ = json_map["name"].(string)

		// password
		inputPassword, ok := json_map["password"].(string)
		if ok {
			correctPassword = inputPassword == password
		}

		// amount
		amount, ok := json_map["amount"].(string)
		if ok {
			amountInt, err = strconv.ParseInt(amount, 10, 64)
			if err != nil {
				fmt.Println("Error parsing amount")
				return realPlayer, correctPassword, amountInt, err
			}
		}

		return realPlayer, correctPassword, amountInt, nil
	}
	cleanName := func(name string) string {
		return strings.ToLower(strings.TrimSpace(name))
	}
	playersList := func() []Player {
		Lock.Lock()
		defer Lock.Unlock()

		players := make([]Player, 0)
		for _, player := range playerData {
			players = append(players, player)
		}
		return players
	}

	buzzChan := make(chan bool)

	e := echo.New()

	// static files
	e.Static("/", "public")

	// pages
	e.GET("/", controllers.Home)

	e.GET("/play/:name", controllers.Play)

	e.GET("/leaderboard", controllers.Leaderboard)
	e.GET("/buzzed-in", controllers.BuzzedIn)
	e.GET("/stats", controllers.Stats)
	e.GET("/host", controllers.Host)
	e.GET("/control", controllers.Control)

	e.GET("/question-number", func(c echo.Context) error {
		Lock.Lock()
		defer Lock.Unlock()
		return c.String(200, fmt.Sprintf("%v", questionNumber))
	})
	e.GET("/players", func(c echo.Context) error {
		Lock.Lock()
		defer Lock.Unlock()
		players := make([]string, 0)
		for _, player := range playerData {
			players = append(players, player.Name)
		}
		// sort alphabetically without case sensitivity
		sort.Slice(players, func(i, j int) bool {
			return cleanName(players[i]) < cleanName(players[j])
		})
		return c.JSON(200, players)
	})

	// player actions
	e.POST("/check-in", func(c echo.Context) error {
		realPlayer, _, _, err := parseJSON(c)
		if err != nil {
			fmt.Println("Error decoding json")
			return err
		}
		Lock.Lock()
		defer Lock.Unlock()

		playerName := cleanName(realPlayer)

		// create new player if not exists
		if _, ok := playerData[playerName]; !ok {
			playerData[playerName] = Player{
				Name:             realPlayer,
				Score:            0,
				CorrectQuestions: make([]string, 0),
				LastUpdate:       time.Now(),
				BuzzIn:           time.Time{},
			}
		}

		return c.String(200, fmt.Sprintf("%v", questionNumber))
	})
	e.POST("/buzz", func(c echo.Context) error {
		realPlayer, _, _, err := parseJSON(c)
		if err != nil {
			fmt.Println("Error decoding json")
			return err
		}

		Lock.Lock()
		defer Lock.Unlock()

		playerName := cleanName(realPlayer)

		var player Player
		player, ok := playerData[playerName]
		if ok {
			// if existing player
			player.BuzzIn = time.Now()
		} else {
			// create new player if not exists
			player = Player{
				Name:             realPlayer,
				Score:            0,
				CorrectQuestions: make([]string, 0),
				LastUpdate:       time.Now(),
				BuzzIn:           time.Now(),
			}
		}
		playerData[playerName] = player

		buzzChan <- true

		return c.String(200, fmt.Sprintf("%v", questionNumber))
	})

	// host actions
	e.POST("/clear", func(c echo.Context) error {
		_, correctPassword, _, err := parseJSON(c)
		if err != nil {
			fmt.Println("Error decoding json")
			return err
		}

		Lock.Lock()
		defer Lock.Unlock()

		// verify password
		if !correctPassword {
			fmt.Println("Unauthorized")
			return c.String(401, "Unauthorized")
		}

		fmt.Println("Clear")
		// clear all player buzz in times
		for playerName, player := range playerData {
			player.BuzzIn = time.Time{}
			playerData[playerName] = player
		}

		return c.String(200, "Clear")
	})
	changeQuestion := func(c echo.Context, inc bool) error {
		_, correctPassword, _, err := parseJSON(c)
		if err != nil {
			fmt.Println("Error decoding json")
			return c.String(400, "Bad Request: Invalid JSON")
		}

		Lock.Lock()
		defer Lock.Unlock()

		// verify password
		if !correctPassword {
			fmt.Println("Unauthorized")
			return c.String(401, "Unauthorized")
		}

		// increment or decrement question number
		if inc {
			questionNumber += 1
		} else {
			questionNumber -= 1
		}

		// clear all player buzz in times
		for _, player := range playerData {
			player.LastUpdate = time.Now()
		}

		return c.String(200, fmt.Sprintf("%v", questionNumber))
	}
	e.POST("/next", func(c echo.Context) error {
		return changeQuestion(c, true)
	})
	e.POST("/prev", func(c echo.Context) error {
		return changeQuestion(c, false)
	})
	e.PUT("/update-score", func(c echo.Context) error {
		realPlayer, correctPassword, amountInt, err := parseJSON(c)
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
		playerName := cleanName(realPlayer)

		Lock.Lock()
		defer Lock.Unlock()

		player, ok := playerData[playerName]
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
			player.CorrectQuestions = append(player.CorrectQuestions, fmt.Sprintf("%d", questionNumber))
		} else {
			player.CorrectQuestions = append(player.CorrectQuestions, fmt.Sprintf("-%d", questionNumber))
		}

		playerData[playerName] = player

		return c.String(200, fmt.Sprintf("%v", player.Score))
	})
	e.DELETE("/delete-player", func(c echo.Context) error {
		realPlayer, correctPassword, _, err := parseJSON(c)
		if err != nil {
			fmt.Println("Error decoding json")
			return c.String(400, "Bad Request: Invalid JSON")
		}

		Lock.Lock()
		defer Lock.Unlock()

		// verify password
		if !correctPassword {
			fmt.Println("Unauthorized")
			return c.String(401, "Unauthorized")
		}

		// verify player exists
		playerName := cleanName(realPlayer)
		if _, ok := playerData[playerName]; !ok {
			fmt.Println("Player not found")
			return c.String(400, "Bad Request: Player not found")
		}

		// delete player
		delete(playerData, playerName)

		return c.String(200, "Player deleted")
	})

	// game information
	e.POST("/leaderboard", func(c echo.Context) error {
		// list of all players
		players := playersList()

		// sort players by score, then by last update time
		sort.Slice(players, func(i, j int) bool {
			if players[i].Score == players[j].Score {
				return players[i].LastUpdate.Before(players[j].LastUpdate)
			}
			return players[i].Score > players[j].Score
		})

		// list of all players and their scores
		playersWithScores := make([][]string, 0)
		for _, player := range players {
			playersWithScores = append(playersWithScores, []string{player.Name, fmt.Sprintf("%d", player.Score)})
		}

		return c.JSON(200, playersWithScores)
	})
	sendBuzzedIn := func(conn *websocket.Conn) {
		// list of all players
		players := playersList()

		// sort players by buzz in time, then alphabetically
		sort.Slice(players, func(i, j int) bool {
			if players[i].BuzzIn.Before(players[j].BuzzIn) {
				return true
			} else if players[i].BuzzIn.After(players[j].BuzzIn) {
				return false
			} else {
				return players[i].Name < players[j].Name
			}
		})

		// list of all players and their buzz in times
		playersWithBuzzIn := make([][]string, 0)
		for _, player := range players {
			if player.BuzzIn.IsZero() {
				continue
			}
			playersWithBuzzIn = append(playersWithBuzzIn, []string{player.Name, player.BuzzIn.Format("03:04:05.000 PM")})
		}

		err := conn.WriteJSON(playersWithBuzzIn)
		if err != nil {
			fmt.Println("Error writing to websocket")
		}
	}
	e.GET("/buzzed-in", func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			fmt.Println("Error upgrading to websocket")
			return err
		}
		defer conn.Close()

		// send initial list of players
		sendBuzzedIn(conn)

		// send when new data is available
		for _ = range buzzChan {
			sendBuzzedIn(conn)
		}

		return nil
	})
	e.POST("/stats", func(c echo.Context) error {
		// list of all players
		players := playersList()

		// sort players by score, then by last update time
		sort.Slice(players, func(i, j int) bool {
			if players[i].Score == players[j].Score {
				return players[i].LastUpdate.Before(players[j].LastUpdate)
			}
			return players[i].Score > players[j].Score
		})

		// list of all players and their scores and correct answers
		playersWithStats := make([][]string, 0)
		for _, player := range players {
			playersWithStats = append(playersWithStats, []string{player.Name, fmt.Sprintf("%d", player.Score), strings.Trim(strings.Join(player.CorrectQuestions, ","), "[]")})
		}

		return c.JSON(200, playersWithStats)
	})

	e.RouteNotFound("/*", controllers.NotFound)

	e.Logger.Fatal(e.Start(":3000"))
}
