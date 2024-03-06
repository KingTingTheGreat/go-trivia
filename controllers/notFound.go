package controllers

import "github.com/labstack/echo/v4"

// not found redirects to home
func NotFound(c echo.Context) error {
	return c.Redirect(302, "/")
}
