package controllers

import (
	"go-trivia/templates"
	"net/http"

	"github.com/labstack/echo/v4"
)

func Stats(c echo.Context) error {
	return render(c, http.StatusOK, templates.Stats(), "Stats | Trivia")
}
