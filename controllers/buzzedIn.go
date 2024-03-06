package controllers

import (
	"go-trivia/templates"
	"net/http"

	"github.com/labstack/echo/v4"
)

func BuzzedIn(c echo.Context) error {
	return render(c, http.StatusOK, templates.BuzzedIn(), "Buzzed In | Trivia")
}
