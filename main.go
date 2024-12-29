package main

import (
	"errors"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/valyala/fasthttp"
	"log/slog"
	"net/http"
	"os"
)

const (
	mixpeekBaseUrl = "https://api/mixpeek.com/"
)

var mixpeekApiKey = os.Getenv("MIXPEEK_API_KEY")

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", hello)

	if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}

func hello(c echo.Context) error {
	return c.String(fasthttp.StatusOK, "hello world!")
}
