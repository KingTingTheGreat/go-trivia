package utils

import (
	"encoding/json"
	"fmt"
	"go-trivia/configs"
	"strconv"

	"github.com/labstack/echo/v4"
)

var Password = configs.Password()

func ParseJSON(c echo.Context) (string, bool, int64, error) {
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
		correctPassword = inputPassword == Password
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
