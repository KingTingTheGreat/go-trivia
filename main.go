package main

import (
	"fmt"
	"go-trivia/controllers"
	"go-trivia/shared"
	"go-trivia/types"
	"go-trivia/utils"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	// static files
	e.Static("/", "public")

	// pages
	e.GET("/", controllers.GetHome)

	e.GET("/play/:name", controllers.GetPlay)

	e.GET("/leaderboard", controllers.GetLeaderboard)
	e.GET("/buzzed-in", controllers.GetBuzzedIn)
	e.GET("/stats", controllers.GetStats)
	e.GET("/host", controllers.GetHost)
	e.GET("/control", controllers.GetControl)

	e.GET("/question-number", func(c echo.Context) error {
		shared.Lock.RLock()
		defer shared.Lock.RUnlock()
		return c.String(200, fmt.Sprintf("%v", shared.QuestionNumber))
	})
	e.GET("/players", func(c echo.Context) error {
		shared.Lock.RLock()
		defer shared.Lock.RUnlock()
		players := make([]string, 0)
		for _, player := range shared.PlayerData {
			players = append(players, player.Name)
		}
		// sort alphabetically without case sensitivity
		sort.Slice(players, func(i, j int) bool {
			return utils.CleanName(players[i]) < utils.CleanName(players[j])
		})
		return c.JSON(200, players)
	})

	// player actions
	e.POST("/check-in", func(c echo.Context) error {
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

		}

		return c.String(200, fmt.Sprintf("%v", shared.QuestionNumber))
	})
	e.POST("/buzz", func(c echo.Context) error {
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
	})

	// host actions
	e.POST("/clear", func(c echo.Context) error {
		shared.Lock.Lock()
		defer shared.Lock.Unlock()

		_, correctPassword, _, err := utils.ParseJSON(c)
		if err != nil {
			fmt.Println("Error decoding json")
			return err
		}

		// verify password
		if !correctPassword {
			fmt.Println("Unauthorized")
			return c.String(401, "Unauthorized")
		}

		fmt.Println("Clear")
		// clear all player buzz in times
		for playerName, player := range shared.PlayerData {
			player.BuzzIn = time.Time{}
			shared.PlayerData[playerName] = player
		}

		shared.BuzzChan <- true

		return c.String(200, "Clear")
	})
	changeQuestion := func(c echo.Context, inc bool) error {
		_, correctPassword, _, err := utils.ParseJSON(c)
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
		for _, player := range shared.PlayerData {
			player.LastUpdate = time.Now()
		}

		return c.String(200, fmt.Sprintf("%v", shared.QuestionNumber))
	}
	e.POST("/next", func(c echo.Context) error {
		return changeQuestion(c, true)
	})
	e.POST("/prev", func(c echo.Context) error {
		return changeQuestion(c, false)
	})
	e.PUT("/update-score", func(c echo.Context) error {
		realPlayer, correctPassword, amountInt, err := utils.ParseJSON(c)
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
		playerName := utils.CleanName(realPlayer)

		shared.Lock.Lock()
		defer shared.Lock.Unlock()

		player, ok := shared.PlayerData[playerName]
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
			player.CorrectQuestions = append(player.CorrectQuestions, fmt.Sprintf("%d", shared.QuestionNumber))
		} else {
			player.CorrectQuestions = append(player.CorrectQuestions, fmt.Sprintf("-%d", shared.QuestionNumber))
		}

		shared.PlayerData[playerName] = player

		shared.LeaderboardChan <- true

		return c.String(200, fmt.Sprintf("%v", player.Score))
	})
	e.DELETE("/delete-player", func(c echo.Context) error {
		realPlayer, correctPassword, _, err := utils.ParseJSON(c)
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

		// verify player exists
		playerName := utils.CleanName(realPlayer)
		if _, ok := shared.PlayerData[playerName]; !ok {
			fmt.Println("Player not found")
			return c.String(400, "Bad Request: Player not found")
		}

		// delete player
		delete(shared.PlayerData, playerName)

		shared.LeaderboardChan <- true

		return c.String(200, "Player deleted")
	})

	// game information
	sendLeaderboard := func(conn *websocket.Conn) {
		fmt.Println("sendLeaderboard")
		// list of all players
		players := utils.PlayersList()

		shared.Lock.RLock()
		// defer shared.Lock.RUnlock()

		// sort players by score, then by last update time
		sort.Slice(players, func(i, j int) bool {
			if players[i].Score == players[j].Score {
				return players[i].LastUpdate.Before(players[j].LastUpdate)
			}
			return players[i].Score > players[j].Score
		})

		shared.Lock.RUnlock()

		// list of all players and their scores
		playersWithScores := make([][]string, 0)
		for _, player := range players {
			playersWithScores = append(playersWithScores, []string{player.Name, fmt.Sprintf("%d", player.Score)})
		}

		err := conn.WriteJSON(playersWithScores)
		if err != nil {
			fmt.Println("Error writing to leaderboard websocket")
		}
	}
	e.GET("/leaderboard-ws", func(c echo.Context) error {
		conn, err := shared.Upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			fmt.Println("Error upgrading to leaderboard websocket")
			return err
		}
		defer conn.Close()

		// send initial list of players
		fmt.Println("sending initial leaderboard")
		sendLeaderboard(conn)

		// send when new data is available
		for _ = range shared.LeaderboardChan {
			fmt.Println("got leaderboardChan data")
			go sendLeaderboard(conn)
		}

		return nil
	})
	sendBuzzedIn := func(conn *websocket.Conn) {
		fmt.Println("send buzzed in")

		// list of all players
		players := utils.PlayersList()

		shared.Lock.RLock()
		defer shared.Lock.RUnlock()

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
	e.GET("/buzzed", func(c echo.Context) error {
		conn, err := shared.Upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			fmt.Println("Error upgrading to websocket")
			return err
		}
		defer conn.Close()

		// send initial list of players
		fmt.Println("sending initial buzzed in")
		sendBuzzedIn(conn)

		// send when new data is available
		for _ = range shared.BuzzChan {
			fmt.Println("BuzzChan update")
			go sendBuzzedIn(conn)
		}

		return nil
	})
	e.POST("/stats", func(c echo.Context) error {
		// list of all players
		players := utils.PlayersList()

		shared.Lock.RLock()
		defer shared.Lock.RUnlock()

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

	e.RouteNotFound("/*", controllers.GetNotFound)

	e.Logger.Fatal(e.Start(":3000"))
}
