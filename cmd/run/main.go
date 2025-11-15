package main

import (
	"context"
	"log"

	"avito-tech/internal/app"
)

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to create a new app: %s", err.Error())
	}

	if err = application.Run(context.Background()); err != nil {
		log.Fatalf("Failed to start the app: %s", err.Error())
	}
}
