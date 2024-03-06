package controllers

import (
	"go-trivia/templates"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/context"
)

func render(ctx echo.Context, status int, t templ.Component, title string) error {
	ctx.Response().Writer.WriteHeader(status)

	err := templates.Layout(t, title).Render(context.Background(), ctx.Response().Writer)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, "failed to render response template")
	}
	return nil
}
