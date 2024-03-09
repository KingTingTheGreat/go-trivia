package controllers

import (
	"fmt"
	"go-trivia/shared"

	"github.com/labstack/echo/v4"
)

func GetQuestionNumber(c echo.Context) error {
	shared.Lock.RLock()
	defer shared.Lock.RUnlock()
	return c.String(200, fmt.Sprintf("%v", shared.QuestionNumber))
}
