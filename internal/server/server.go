package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Prysya/go-final-project/internal/handlers"
	"github.com/Prysya/go-final-project/pkg/db"
	"github.com/Prysya/go-final-project/pkg/repository"
)

type Server struct {
	logger *log.Logger
	server *http.Server
}

type Config struct {
	Port   string
	DBFile string
}

func getConfig() Config {
	config := Config{}

	if port := os.Getenv("TODO_PORT"); port != "" {
		if _, err := strconv.Atoi(port); err == nil {
			config.Port = ":" + port
		} else {
			config.Port = ":7540"
		}
	} else {
		config.Port = ":7540"
	}

	if dbFile := os.Getenv("TODO_DBFILE"); dbFile != "" {
		config.DBFile = dbFile
	} else {
		config.DBFile = "scheduler.db"
	}

	return config
}

func CreateServer(logger *log.Logger) *Server {
	config := getConfig()

	if err := db.Init(config.DBFile); err != nil {
		logger.Fatalf("Ошибка инициализации базы данных: %v", err)
	}

	router := http.NewServeMux()
	taskRepo, err := repository.NewTaskRepository()
	if err != nil {
		log.Fatal(err)
	}

	router.Handle("/", handlers.GetFileServerHandler())
	router.HandleFunc("/api/nextdate", handlers.NextDateHandler)
	router.HandleFunc("/api/task", handlers.TaskHandler(taskRepo))
	router.HandleFunc("/api/task/done", handlers.TaskDoneHandler(taskRepo))
	router.HandleFunc("/api/tasks", handlers.TasksHandler(taskRepo))

	httpServer := &http.Server{
		Addr:         config.Port,
		Handler:      router,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return &Server{
		logger: logger,
		server: httpServer,
	}
}

func (s *Server) Start() error {
	s.logger.Printf("Сервер запущен по адресу: http://localhost%s", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Println("Останавливаем сервер...")

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Printf("Ошибка при остановке HTTP сервера: %v", err)
		s.server.Close()
	}

	if err := db.Close(); err != nil {
		s.logger.Printf("Ошибка при закрытии БД: %v", err)
		return err
	}

	return nil
}
