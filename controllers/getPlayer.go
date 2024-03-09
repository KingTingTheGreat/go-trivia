package controllers

import (
	"fmt"
	"go-trivia/templates"
	"net/http"

	"github.com/labstack/echo/v4"
)

func GetPlay(c echo.Context) error {
	name := c.Param("name")
	return render(c, http.StatusOK, templates.Player(name), fmt.Sprintf("%s | Trivia", name))
}
