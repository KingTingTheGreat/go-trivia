package main

import (
	"encoding/json"
	"fmt"
	"go-trivia/configs"
	"go-trivia/controllers"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

func main() {
	// prevent data races
	Lock := &sync.Mutex{}

	// players and when they first buzzed in this round
	playerTimes := make(map[string]time.Time)
	// real username
	playerNames := make(map[string]string)
	// players and their scores
	playerScores := make(map[string]int64)
	// players and how many questions they have answered correctly
	playerCorrect := make(map[string][]string)
	// players and the last time their score was updated
	// leaderboard is sorted by score, then by last update
	lastUpdate := make(map[string]time.Time)

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
		realPlayer = json_map["name"].(string)

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
		for playerName, _ := range playerScores {
			players = append(players, playerNames[playerName])
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

		// set player name to real name if not exists
		if _, ok := playerNames[playerName]; !ok {
			playerNames[playerName] = realPlayer
		}

		// set player score to 0 if not exists
		if _, ok := playerScores[playerName]; !ok {
			playerScores[playerName] = 0
			playerCorrect[playerName] = make([]string, 0)
			lastUpdate[playerName] = time.Now()
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

		// player buzzed in
		if _, ok := playerTimes[playerName]; !ok {
			playerTimes[playerName] = time.Now()
		}

		// store real name if not exists
		if _, ok := playerNames[playerName]; !ok {
			playerNames[playerName] = realPlayer
		}

		// give player a score if this is their first buzz
		if _, ok := playerScores[playerName]; !ok {
			playerScores[playerName] = 0
			playerCorrect[playerName] = make([]string, 0)
			lastUpdate[playerName] = time.Now()
		}

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

		// clear all player buzz in times
		playerTimes = make(map[string]time.Time)
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
		// reset all player buzz in times
		playerTimes = make(map[string]time.Time)
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

		Lock.Lock()
		defer Lock.Unlock()

		// verify password
		if !correctPassword {
			fmt.Println("Unauthorized")
			return c.String(401, "Unauthorized")
		}
		// verify playername and amount
		playerName := cleanName(realPlayer)

		oldScore, ok := playerScores[playerName]
		if !ok {
			fmt.Println("Player not found")
			return c.String(400, "Bad Request: Player not found")
		}

		// update player score
		playerScores[playerName] = oldScore + amountInt
		lastUpdate[playerName] = time.Now()

		// update player correct
		if amountInt > 0 {
			playerCorrect[playerName] = append(playerCorrect[playerName], fmt.Sprintf("%d", questionNumber))
			// } else if len(playerCorrect[playerName]) > 0 {
			// playerCorrect[playerName] = playerCorrect[playerName][:len(playerCorrect[playerName])-1]
		} else {
			playerCorrect[playerName] = append(playerCorrect[playerName], fmt.Sprintf("-%d", questionNumber))

		}

		return c.String(200, fmt.Sprintf("%v", playerScores[playerName]))
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
		if _, ok := playerScores[playerName]; !ok {
			fmt.Println("Player not found")
			return c.String(400, "Bad Request: Player not found")
		}

		// delete player
		delete(playerScores, playerName)
		delete(playerCorrect, playerName)
		delete(playerTimes, playerName)
		delete(playerNames, playerName)
		delete(lastUpdate, playerName)

		return c.String(200, "Player deleted")
	})

	// game information
	e.POST("/leaderboard", func(c echo.Context) error {
		Lock.Lock()
		defer Lock.Unlock()

		// list of all players and their scores
		playerWithScores := make([][]string, 0)
		for playerName, score := range playerScores {
			playerWithScores = append(playerWithScores, []string{playerNames[playerName], fmt.Sprintf("%d", score)})
		}

		// sort players by score, then by last update time
		sort.Slice(playerWithScores, func(i, j int) bool {
			a := playerScores[playerWithScores[i][0]]
			b := playerScores[playerWithScores[j][0]]
			if a == b {
				return lastUpdate[playerWithScores[i][0]].Before(lastUpdate[playerWithScores[j][0]])
			}
			return a > b
		})

		return c.JSON(200, playerWithScores)
	})
	e.POST("/buzzed-in", func(c echo.Context) error {
		Lock.Lock()
		defer Lock.Unlock()

		// list all players and their buzz in times, in order of buzz in
		players := make([][]string, 0)
		for playerName, _ := range playerTimes {
			players = append(players, []string{playerNames[playerName], playerTimes[playerName].Format("03:04:05.000 PM")})
		}
		sort.Slice(players, func(i, j int) bool {
			return playerTimes[players[i][0]].Before(playerTimes[players[j][0]])
		})

		return c.JSON(200, players)
	})
	e.POST("/stats", func(c echo.Context) error {
		Lock.Lock()
		defer Lock.Unlock()

		// list all players and their scores and correct answers
		players := make([][]string, 0)
		for playerName, score := range playerScores {
			players = append(players, []string{playerNames[playerName], fmt.Sprintf("%d", score), strings.Trim(strings.Join(playerCorrect[playerName], ","), "[]")})
		}

		// sort players by score, then by last update time
		sort.Slice(players, func(i, j int) bool {
			a := playerScores[players[i][0]]
			b := playerScores[players[j][0]]
			if a == b {
				return lastUpdate[players[i][0]].Before(lastUpdate[players[j][0]])
			}
			return a > b
		})

		return c.JSON(200, players)
	})

	e.RouteNotFound("/*", controllers.NotFound)

	e.Logger.Fatal(e.Start(":3000"))
}
