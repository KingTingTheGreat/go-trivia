package controllers

import "github.com/labstack/echo/v4"

// not found redirects to home
func GetNotFound(c echo.Context) error {
	return c.Redirect(302, "/")
}
