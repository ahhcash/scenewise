package main

import (
	_ "github.com/joho/godotenv/autoload"
	"os"
)

const (
	mixpeekBaseUrl = "https://api/mixpeek.com/"
)

var mixpeekApiKey = os.Getenv("MIXPEEK_API_KEY")

func main() {
	// to make the api
}
