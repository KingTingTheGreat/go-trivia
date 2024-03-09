package controllers

import (
	"go-trivia/templates"
	"net/http"

	"github.com/labstack/echo/v4"
)

func GetLeaderboard(c echo.Context) error {
	return render(c, http.StatusOK, templates.Leaderboard(), "Leaderboard | Trivia")
}
