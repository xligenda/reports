package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/xligenda/reports/internal/app"
)

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	defer application.Shutdown()
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}
