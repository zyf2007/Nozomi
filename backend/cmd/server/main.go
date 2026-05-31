package main

import (
	"flag"
	"log"

	"github.com/joho/godotenv"
	"nozomi-relay/internal/server"
)

func main() {
	webMode := flag.String("web", "auto", "static web mode: auto or off")
	flag.Parse()

	_ = godotenv.Load()
	log.Fatal(server.Run(server.Options{WebMode: *webMode}))
}
