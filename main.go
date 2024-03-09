package main

import (
	"go-trivia/controllers"

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

	e.GET("/question-number", controllers.GetQuestionNumber)
	e.GET("/players", controllers.GetPlayers)

	// player actions
	e.POST("/check-in", controllers.PostCheckIn)
	e.POST("/buzz", controllers.PostBuzz)

	// host actions
	e.POST("/clear", controllers.PostClear)
	e.POST("/next", controllers.PostNext)
	e.POST("/prev", controllers.PostPrev)
	e.PUT("/update-score", controllers.PutUpdateScore)
	e.DELETE("/player", controllers.DeletePlayer)

	e.GET("/leaderboard-ws", controllers.GetLeaderboardWs)

	e.GET("/buzzed-ws", controllers.GetBuzzedInWs)
	e.POST("/stats", controllers.PostStats)

	e.RouteNotFound("/*", controllers.GetNotFound)

	e.Logger.Fatal(e.Start(":3000"))
}
