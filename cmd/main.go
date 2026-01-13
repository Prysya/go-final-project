package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Prysya/go-final-project/internal/server"
)

func main() {
	logger := log.New(os.Stdout, "SERVER: ", log.LstdFlags|log.Lshortfile)

	srv := server.CreateServer(logger)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			logger.Fatalf("Ошибка запуска сервера: %s", err.Error())
		}
	}()

	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		logger.Printf("Ошибка при остановке сервера: %v", err)
	}
}
