package main

import (
	"fmt"
	"github.com/ahhcash/vsearch/api/handlers"
	"github.com/ahhcash/vsearch/config"
	_ "github.com/joho/godotenv/autoload"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/valyala/fasthttp"
)

func main() {
	cfg := config.Load()
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())

	searchHandler := handlers.NewSearchHandler(cfg)

	e.GET("/", hello)
	e.POST("/search", searchHandler.Search)
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(fasthttp.StatusOK, map[string]string{
			"status": "healthy",
		})
	})

	// Updated port binding
	addr := fmt.Sprintf(":%s", cfg.Port)
	e.Logger.Info(fmt.Sprintf("Starting server on %s", addr))
	e.Logger.Fatal(e.Start(addr))
}

func hello(c echo.Context) error {
	return c.String(fasthttp.StatusOK, "hello world!")
}
