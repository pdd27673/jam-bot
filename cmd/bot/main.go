// cmd/bot/main.go
package main

import (
	"fmt"
	"jam-bot/internal/bot"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("local.env")
	if err != nil {
		fmt.Println("[INFO] No .env file found. Proceeding with environment variables.")
	}

	err = bot.StartBot()
	if err != nil {
		fmt.Printf("[ERROR] error starting the bot: %s", err)
	}
}
