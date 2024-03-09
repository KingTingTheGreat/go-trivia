package controllers

import (
	"go-trivia/utils"

	"github.com/labstack/echo/v4"
)

func PostNext(c echo.Context) error {
	return utils.ChangeQuestion(c, true)
}
