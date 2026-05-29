package main

import (
	"log"

	"github.com/joho/godotenv"
	"nozomi-relay/internal/server"
)

func main() {
	_ = godotenv.Load()
	log.Fatal(server.Run())
}
