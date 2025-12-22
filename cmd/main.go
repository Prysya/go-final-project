package main

import (
	"log"
	"os"

	"github.com/Prysya/go-final-project/internal/server"
)

func main() {
	logger := log.New(os.Stdout, "SERVER: ", log.LstdFlags|log.Lshortfile)

	srv := server.CreateServer(logger)

	if err := srv.Start(); err != nil {
		logger.Fatalf("Ошибка запуска сервера: %s", err.Error())
	}

	defer srv.Stop()
}
